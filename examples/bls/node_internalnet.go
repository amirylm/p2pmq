package blstest

import (
	"context"
	"fmt"
	"strconv"

	"github.com/amirylm/p2pmq/proto"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func (n *Node) validateInternalNet(sr SignedReport) proto.ValidationResult {
	share, ok := n.Shares.Get(sr.Network)
	if !ok { // this node is not a signer in this network, ignore
		return proto.ValidationResult_IGNORE
	}
	if !share.Validate(sr) {
		return proto.ValidationResult_REJECT
	}
	reports, ok := n.internalReports.Get(sr.Network)
	if ok && reports.Get(fmt.Sprintf("%d", sr.SignerID), sr.SeqNumber) != nil {
		return proto.ValidationResult_IGNORE
	}
	return proto.ValidationResult_ACCEPT
}

func (n *Node) consumeInternalNet(sr SignedReport) error {
	share, ok := n.Shares.Get(sr.Network)
	if ok {
		n.threadC.Go(func(ctx context.Context) {
			reports, ok := n.internalReports.Get(sr.Network)
			if !ok {
				reports = NewReportBuffer(reportBufferSize)
				n.internalReports.Add(sr.Network, reports)
			}
			if !reports.Add(fmt.Sprintf("%d", sr.SignerID), sr) {
				return
			}
			signers := reports.Lookup(sr.SeqNumber)
			leader := n.getLeader(sr.Network, sr.SeqNumber)
			if n.isProcessable(fmt.Sprintf("%d", share.SignerID), fmt.Sprintf("%d", leader), signers...) {
				share.Sign(&sr)
				reports.Add(fmt.Sprintf("%d", share.SignerID), sr)
				if err := n.Broadcast(ctx, sr); err != nil {
					fmt.Printf("Error broadcasting report: %s\n", err)
					return
				}
				signers = append(signers, fmt.Sprintf("%d", share.SignerID))
			}
			if n.hasQuorum(sr, signers...) {
				fmt.Printf("Signer %d: quorum reached for network %s, seq %d (%d signers)\n", share.SignerID, sr.Network, sr.SeqNumber, len(signers))
				n.threadC.Go(func(ctx context.Context) {
					signatures := n.collectSignatures(sr.SeqNumber, reports, signers...)
					if len(signatures) > n.quorumCount(sr.Network) {
						signed, err := n.prepareReport(signatures, sr)
						if err != nil {
							fmt.Printf("Error preparing report: %s\n", err)
							return
						}
						if err := n.Broadcast(ctx, signed); err != nil {
							fmt.Printf("Error broadcasting report: %s\n", err)
							return
						}
					}
				})
			}
		})
	}
	return nil
}

func (n *Node) prepareReport(signatures map[uint64][]byte, sr SignedReport) (SignedReport, error) {
	fmt.Printf("Reconstructing signature for network %s, seq %d\n", sr.Network, sr.SeqNumber)
	sign, err := ReconstructSigs(signatures)
	if err != nil {
		fmt.Printf("Error reconstructing signature: %s\n", err)
		return sr, err
	}
	sr.SigHex = sign.SerializeToHexStr()
	sr.SignerID = 0 // no signer id as this is a reconstructed signature
	return sr, nil
}

func (n *Node) collectSignatures(seq uint64, reports *ReportBuffer, signers ...string) map[uint64][]byte {
	signatures := make(map[uint64][]byte)
	for _, s := range signers {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing signer id: %s\n", err)
			continue
		}
		rep := reports.Get(s, seq)
		if rep == nil {
			fmt.Printf("Error retrieving report for signer %d\n", id)
			continue
		}
		sign := &bls.Sign{}
		if err := sign.DeserializeHexStr(rep.SigHex); err != nil {
			fmt.Printf("Error deserializing signature: %s\n", err)
			continue
		}
		signatures[id] = sign.Serialize()
	}
	return signatures
}

func (n *Node) hasQuorum(sr SignedReport, signers ...string) bool {
	// count = 3f + 1
	count := n.quorumCount(sr.Network)

	return len(signers) >= count
}

func (n *Node) quorumCount(net string) int {
	share, ok := n.Shares.Get(net)
	if !ok {
		return 0
	}
	return share.QuorumCount()
}

// getLeader returns the leader for the given sequence number (round robin).
func (n *Node) getLeader(net string, seq uint64) uint64 {
	share, ok := n.Shares.Get(net)
	if !ok {
		return 0
	}
	return RoundRobinLeader(seq, share.Signers)
}

// isProcessable ensures that we sign once and only leaders can trigger a new sequence
func (n *Node) isProcessable(opid, leader string, signers ...string) bool {
	var leaderSigned, signed bool
	for _, signer := range signers {
		if leader == signer {
			leaderSigned = true
		}
		if opid == signer {
			signed = true
		}
	}
	return leaderSigned && !signed
}
