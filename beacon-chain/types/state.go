package types

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/prysmaticlabs/prysm/beacon-chain/params"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"golang.org/x/crypto/blake2b"
)

// ActiveState contains fields of current state of beacon chain,
// it changes every block.
type ActiveState struct {
	data *pb.ActiveState
}

// CrystallizedState contains fields of every Slot state,
// it changes every Slot.
type CrystallizedState struct {
	data *pb.CrystallizedState
}

// NewCrystallizedState creates a new crystallized state with a explicitly set data field.
func NewCrystallizedState(data *pb.CrystallizedState) *CrystallizedState {
	return &CrystallizedState{data: data}
}

// NewActiveState creates a new active state with a explicitly set data field.
func NewActiveState(data *pb.ActiveState) *ActiveState {
	return &ActiveState{data: data}
}

// NewGenesisStates initializes a beacon chain with starting parameters.
func NewGenesisStates() (*ActiveState, *CrystallizedState) {
	active := &ActiveState{
		data: &pb.ActiveState{
			PendingAttestations: []*pb.AttestationRecord{},
			RecentBlockHashes:   [][]byte{},
		},
	}
	// We seed the genesis crystallized state with a bunch of validators to
	// bootstrap the system.
	// TODO: Perform this task from some sort of genesis state json config instead.
	var validators []*pb.ValidatorRecord
	for i := 0; i < params.BootstrappedValidatorsCount; i++ {
		validator := &pb.ValidatorRecord{StartDynasty: 0, EndDynasty: params.DefaultEndDynasty, Balance: params.DefaultBalance, WithdrawalAddress: []byte{}, PublicKey: 0}
		validators = append(validators, validator)
	}
	crystallized := &CrystallizedState{
		data: &pb.CrystallizedState{
			LastStateRecalc:        0,
			JustifiedStreak:        0,
			LastJustifiedSlot:      0,
			LastFinalizedSlot:      0,
			CurrentDynasty:         0,
			CrosslinkingStartShard: 0,
			TotalDeposits:          0,
			DynastySeed:            []byte{},
			DynastySeedLastReset:   0,
			CrosslinkRecords:       []*pb.CrosslinkRecord{},
			Validators:             validators,
			IndicesForHeights:      []*pb.ShardAndCommitteeArray{},
		},
	}
	return active, crystallized
}

// NewAttestationRecord initializes an attestation record with default parameters.
func NewAttestationRecord() *pb.AttestationRecord {
	return &pb.AttestationRecord{
		Slot:                0,
		ShardId:             0,
		ObliqueParentHashes: [][]byte{},
		ShardBlockHash:      []byte{0},
		AttesterBitfield:    nil,
		AggregateSig:        []uint64{0, 0},
	}
}

// Proto returns the underlying protobuf data within a state primitive.
func (a *ActiveState) Proto() *pb.ActiveState {
	return a.data
}

// Marshal encodes active state object into the wire format.
func (a *ActiveState) Marshal() ([]byte, error) {
	return proto.Marshal(a.data)
}

// Hash serializes the active state object then uses
// blake2b to hash the serialized object.
func (a *ActiveState) Hash() ([32]byte, error) {
	data, err := proto.Marshal(a.data)
	if err != nil {
		return [32]byte{}, err
	}
	var hash [32]byte
	h := blake2b.Sum512(data)
	copy(hash[:], h[:32])
	return hash, nil
}

// BlockHashForSlot returns the block hash of a given slot given a lowerBound and upperBound.
func (a *ActiveState) BlockHashForSlot(slot uint64, block *Block) ([]byte, error) {
	sback := int(block.SlotNumber()) - params.CycleLength*2
	if !(sback <= int(slot) && int(slot) < sback+params.CycleLength*2) {
		return nil, fmt.Errorf("can not return block hash of a given slot, input slot %v has to be in between %v and %v", slot, sback, sback+params.CycleLength*2)
	}
	if sback < 0 {
		return a.RecentBlockHashes()[slot].Bytes(), nil
	}
	return a.RecentBlockHashes()[int(slot)-sback].Bytes(), nil
}

// PendingAttestations returns attestations that have not yet been processed.
func (a *ActiveState) PendingAttestations() []*pb.AttestationRecord {
	return a.data.PendingAttestations
}

// NewPendingAttestation inserts a new pending attestaton fields.
func (a *ActiveState) NewPendingAttestation(record *pb.AttestationRecord) {
	a.data.PendingAttestations = append(a.data.PendingAttestations, record)
}

// LatestPendingAttestation returns the latest pending attestaton fields.
func (a *ActiveState) LatestPendingAttestation() *pb.AttestationRecord {
	if len(a.data.PendingAttestations) == 0 {
		return nil
	}
	return a.data.PendingAttestations[len(a.data.PendingAttestations)-1]
}

// ClearPendingAttestations clears attestations that have not yet been processed.
func (a *ActiveState) ClearPendingAttestations() {
	for i := range a.data.PendingAttestations {
		a.data.PendingAttestations[i] = &pb.AttestationRecord{}
	}
}

// RecentBlockHashes returns the most recent 2*EPOCH_LENGTH block hashes.
func (a *ActiveState) RecentBlockHashes() []common.Hash {
	var blockhashes []common.Hash
	for _, hash := range a.data.RecentBlockHashes {
		blockhashes = append(blockhashes, common.BytesToHash(hash))
	}
	return blockhashes
}

// ClearRecentBlockHashes resets the most recent 64 block hashes.
func (a *ActiveState) ClearRecentBlockHashes() {
	a.data.RecentBlockHashes = [][]byte{}
}

// Proto returns the underlying protobuf data within a state primitive.
func (c *CrystallizedState) Proto() *pb.CrystallizedState {
	return c.data
}

// Marshal encodes crystallized state object into the wire format.
func (c *CrystallizedState) Marshal() ([]byte, error) {
	return proto.Marshal(c.data)
}

// Hash serializes the crystallized state object then uses
// blake2b to hash the serialized object.
func (c *CrystallizedState) Hash() ([32]byte, error) {
	data, err := proto.Marshal(c.data)
	if err != nil {
		return [32]byte{}, err
	}
	var hash [32]byte
	h := blake2b.Sum512(data)
	copy(hash[:], h[:32])
	return hash, nil
}

// LastStateRecalc returns when the last time crystallized state recalculated.
func (c *CrystallizedState) LastStateRecalc() uint64 {
	return c.data.LastStateRecalc
}

// SetStateRecalc sets last state recalc.
func (c *CrystallizedState) SetStateRecalc(slot uint64) {
	c.data.LastStateRecalc = slot
}

// JustifiedStreak returns number of consecutive justified slots ending at head.
func (c *CrystallizedState) JustifiedStreak() uint64 {
	return c.data.JustifiedStreak
}

// ClearJustifiedStreak clears the number of consecutive justified slots.
func (c *CrystallizedState) ClearJustifiedStreak() {
	c.data.JustifiedStreak = 0
}

// CrosslinkingStartShard returns next shard that crosslinking assignment will start from.
func (c *CrystallizedState) CrosslinkingStartShard() uint64 {
	return c.data.CrosslinkingStartShard
}

// LastJustifiedSlot return the last justified slot of the beacon chain.
func (c *CrystallizedState) LastJustifiedSlot() uint64 {
	return c.data.LastJustifiedSlot
}

// SetLastJustifiedSlot sets the last justified Slot of the beacon chain.
func (c *CrystallizedState) SetLastJustifiedSlot(Slot uint64) {
	c.data.LastJustifiedSlot = Slot
}

// LastFinalizedSlot returns the last finalized Slot of the beacon chain.
func (c *CrystallizedState) LastFinalizedSlot() uint64 {
	return c.data.LastFinalizedSlot
}

// SetLastFinalizedSlot sets last justified Slot of the beacon chain.
func (c *CrystallizedState) SetLastFinalizedSlot(Slot uint64) {
	c.data.LastFinalizedSlot = Slot
}

// CurrentDynasty returns the current dynasty of the beacon chain.
func (c *CrystallizedState) CurrentDynasty() uint64 {
	return c.data.CurrentDynasty
}

// IncrementCurrentDynasty increments current dynasty by one.
func (c *CrystallizedState) IncrementCurrentDynasty() {
	c.data.CurrentDynasty++
}

// TotalDeposits returns total balance of deposits.
func (c *CrystallizedState) TotalDeposits() uint64 {
	return c.data.TotalDeposits
}

// SetTotalDeposits sets total balance of deposits.
func (c *CrystallizedState) SetTotalDeposits(total uint64) {
	c.data.TotalDeposits = total
}

// DynastySeed is used to select the committee for each shard.
func (c *CrystallizedState) DynastySeed() [32]byte {
	var h [32]byte
	copy(h[:], c.data.DynastySeed)
	return h
}

// DynastySeedLastReset is the last finalized Slot that the crosslink seed was reset.
func (c *CrystallizedState) DynastySeedLastReset() uint64 {
	return c.data.DynastySeedLastReset
}

// Validators returns list of validators.
func (c *CrystallizedState) Validators() []*pb.ValidatorRecord {
	return c.data.Validators
}

// ValidatorsLength returns the number of total validators.
func (c *CrystallizedState) ValidatorsLength() int {
	return len(c.data.Validators)
}

// SetValidators sets the validator set.
func (c *CrystallizedState) SetValidators(validators []*pb.ValidatorRecord) {
	c.data.Validators = validators
}

// IndicesForHeights returns what active validators are part of the attester set
// at what height, and in what shard.
func (c *CrystallizedState) IndicesForHeights() []*pb.ShardAndCommitteeArray {
	return c.data.IndicesForHeights
}

// ClearIndicesForHeights clears the IndicesForHeights set.
func (c *CrystallizedState) ClearIndicesForHeights() {
	c.data.IndicesForHeights = []*pb.ShardAndCommitteeArray{}
}

// CrosslinkRecords returns records about the most recent cross link or each shard.
func (c *CrystallizedState) CrosslinkRecords() []*pb.CrosslinkRecord {
	return c.data.CrosslinkRecords
}

// UpdateJustifiedSlot updates the justified and finalized Slot during a dynasty transition.
func (c *CrystallizedState) UpdateJustifiedSlot(currentSlot uint64) {
	slot := c.LastJustifiedSlot()
	c.SetLastJustifiedSlot(currentSlot)

	if currentSlot == (slot + 1) {
		c.SetLastFinalizedSlot(slot)
	}
}
