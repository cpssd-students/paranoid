package pnetclient

import (
	"fmt"
	"log"

	"github.com/cpssd-students/paranoid/cmd/pfsd/globals"
	"github.com/cpssd-students/paranoid/cmd/pfsd/keyman"
)

// Distribute chunked key over the network
func Distribute(key *keyman.Key, peers []globals.Node, generation int) error {
	numPieces := int64(len(peers) + 1)
	requiredPieces := numPieces/2 + 1
	log.Print("Generating pieces.")
	pieces, err := keyman.GeneratePieces(key, numPieces, requiredPieces)
	if err != nil {
		log.Printf("Could not chunk key: %v", err)
		return fmt.Errorf("could not chunk key: %w", err)
	}
	// We always keep the first piece and distribute the rest
	_ = globals.HeldKeyPieces.AddPiece(int64(generation), globals.ThisNode.UUID, pieces[0])
	count := int64(1)

	for i := 1; i < len(pieces); i++ {
		err := SendKeyPiece(peers[i-1].UUID, int64(generation), pieces[i], false)
		if err != nil {
			log.Printf("Error sending key piece: %v", err)
		} else {
			count++
		}
	}

	if count >= requiredPieces {
		if err := globals.RaftNetworkServer.RequestOwnerComplete(
			globals.ThisNode.UUID,
			int64(generation),
		); err != nil {
			log.Printf("Error marking generation complete: %v", err)
		} else {
			log.Printf("Successfully completed generation %d", generation)
		}
	}
	return nil
}

// KSMObserver does raft stuff when the key state machine changes
func KSMObserver(ksm *keyman.KeyStateMachine) {
	defer globals.Wait.Done()
	for {
		select {
		case _, ok := <-globals.Quit:
			if !ok {
				return
			}
		case <-ksm.Events:
		replicationLoop:
			for {
				select {
				case <-ksm.Events:
					//Keep this channel clear
				default:
					done := true
					for g := ksm.GetCurrentGeneration(); g <= ksm.GetInProgressGenertion(); g++ {
						if !ksm.NeedsReplication(globals.ThisNode.UUID, g) {
							continue
						}
						done = false
						nodes, err := ksm.GetNodes(g)
						if err != nil {
							log.Printf("Unable to get nodes for generation %d: %v", g, err)
							continue
						}
						var peers []globals.Node
						for _, v := range nodes {
							if v != globals.ThisNode.UUID {
								globalNode, err := globals.Nodes.GetNode(v)
								if err != nil {
									log.Printf("Unable to lookup node %s: %s", v, err)
									peers = append(peers, globals.Node{})
								} else {
									peers = append(peers, globalNode)
								}
							}
						}
						_ = Distribute(globals.EncryptionKey, peers, int(g))
					}
					for i := int64(0); i < ksm.GetCurrentGeneration(); i++ {
						err := globals.HeldKeyPieces.DeleteGeneration(i)
						if err != nil {
							log.Printf("Unable to delete generation: %v", err)
						}
					}
					if done {
						break replicationLoop
					}
				}
			}
		}
	}
}
