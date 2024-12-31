package axongo

import (
	"fmt"
	"sort"

	"github.com/axonfibre/fibre.go/constraints"
	"github.com/axonfibre/fibre.go/ierrors"
	"github.com/axonfibre/fibre.go/serializer/v2"
)

var (
	// ErrNonUniqueUnlockConditions gets returned when multiple UnlockCondition(s) with the same UnlockConditionType exist within sets.
	ErrNonUniqueUnlockConditions = ierrors.New("non unique unlock conditions within outputs")
	// ErrTimelockNotExpired gets returned when timelocks in a UnlockConditionSet are not expired.
	ErrTimelockNotExpired = ierrors.New("timelock not expired")
	// ErrExpirationConditionZero gets returned when an ExpirationUnlockCondition has set the slot index to zero.
	ErrExpirationConditionZero = ierrors.New("expiration condition is zero")
	// ErrTimelockConditionZero gets returned when a TimelockUnlockCondition has set the slot index to zero.
	ErrTimelockConditionZero = ierrors.New("timelock condition is zero")
	// ErrTimelockCommitmentInputMissing gets returned when a TX containing a TimelockUnlockCondition
	// does not have a commitment input.
	ErrTimelockCommitmentInputMissing = ierrors.New("transaction's containing a timelock condition require a commitment input")
	// ErrExpirationCommitmentInputMissing gets returned when a TX containing an ExpirationUnlockCondition
	// does not have a commitment input.
	ErrExpirationCommitmentInputMissing = ierrors.New("transaction's containing an expiration condition require a commitment input")
	// ErrExpirationNotUnlockable gets returned when a ExpirationUnlockCondition cannot be unlocked by return address or the owner.
	ErrExpirationNotUnlockable = ierrors.New("expiration unlock condition cannot be unlocked")
)

// UnlockConditionType defines the type of UnlockCondition.
type UnlockConditionType byte

const (
	// UnlockConditionAddress denotes an AddressUnlockCondition.
	UnlockConditionAddress UnlockConditionType = iota
	// UnlockConditionStorageDepositReturn denotes a StorageDepositReturnUnlockCondition.
	UnlockConditionStorageDepositReturn
	// UnlockConditionTimelock denotes a TimelockUnlockCondition.
	UnlockConditionTimelock
	// UnlockConditionExpiration denotes an ExpirationUnlockCondition.
	UnlockConditionExpiration
	// UnlockConditionStateControllerAddress denotes a StateControllerAddressUnlockCondition.
	UnlockConditionStateControllerAddress
	// UnlockConditionGovernorAddress denotes a GovernorAddressUnlockCondition.
	UnlockConditionGovernorAddress
	// UnlockConditionImmutableAccount denotes an ImmutableAccountUnlockCondition.
	UnlockConditionImmutableAccount
)

func (unlockCondType UnlockConditionType) String() string {
	if int(unlockCondType) >= len(unlockCondNames) {
		return fmt.Sprintf("unknown unlock condition type: %d", unlockCondType)
	}

	return unlockCondNames[unlockCondType]
}

var (
	unlockCondNames = [UnlockConditionImmutableAccount + 1]string{
		"AddressUnlockCondition",
		"StorageDepositReturnUnlockCondition",
		"TimelockUnlockCondition",
		"ExpirationUnlockCondition",
		"StateControllerAddressUnlockCondition",
		"GovernorAddressUnlockCondition",
		"ImmutableAccountUnlockCondition",
	}
)

// UnlockCondition is an abstract building block defining the unlock conditions of an Output.
type UnlockCondition interface {
	Sizer
	NonEphemeralObject
	ProcessableObject
	constraints.Cloneable[UnlockCondition]
	constraints.Equalable[UnlockCondition]
	constraints.Comparable[UnlockCondition]

	// Type returns the type of the UnlockCondition.
	Type() UnlockConditionType
}

// UnlockConditions is a slice of UnlockCondition(s).
type UnlockConditions[T UnlockCondition] []T

func (f UnlockConditions[T]) Equal(other UnlockConditions[T]) bool {
	if len(f) != len(other) {
		return false
	}

	for idx, unlockCondition := range f {
		if !unlockCondition.Equal(other[idx]) {
			return false
		}
	}

	return true
}

func (f UnlockConditions[T]) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	var sumCost StorageScore
	for _, unlockCond := range f {
		sumCost += unlockCond.StorageScore(storageScoreStruct, nil)
	}

	return sumCost
}

func (f UnlockConditions[T]) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var workScoreUnlockConds WorkScore
	for _, unlockCond := range f {
		workScoreUnlockCond, err := unlockCond.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}

		workScoreUnlockConds, err = workScoreUnlockConds.Add(workScoreUnlockCond)
		if err != nil {
			return 0, err
		}
	}

	return workScoreUnlockConds, nil
}

// Clone clones the UnlockConditions.
func (f UnlockConditions[T]) Clone() UnlockConditions[T] {
	cpy := make(UnlockConditions[T], len(f))
	for i, v := range f {
		//nolint:forcetypeassert // we can safely assume that this is of type T
		cpy[i] = v.Clone().(T)
	}

	return cpy
}

func (f UnlockConditions[T]) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, uc := range f {
		sum += uc.Size()
	}

	return sum
}

// Set converts the slice into an UnlockConditionSet.
// Returns an error if an UnlockConditionType occurs multiple times.
func (f UnlockConditions[T]) Set() (UnlockConditionSet, error) {
	set := make(UnlockConditionSet)
	for _, block := range f {
		if _, has := set[block.Type()]; has {
			return nil, ErrNonUniqueUnlockConditions
		}
		set[block.Type()] = block
	}

	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that an UnlockConditions slice does not contain the same UnlockConditionType multiple times.
func (f UnlockConditions[T]) MustSet() UnlockConditionSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}

	return set
}

// Upsert adds the given unlock condition or updates the previous one if existing.
func (f *UnlockConditions[T]) Upsert(unlockCondition T) {
	for i, ele := range *f {
		if ele.Type() == unlockCondition.Type() {
			(*f)[i] = unlockCondition

			return
		}
	}
	*f = append(*f, unlockCondition)
}

// Remove removes the given unlock condition if it exists.
func (f *UnlockConditions[T]) Remove(unlockConditionType UnlockConditionType) bool {
	for i, ele := range *f {
		if ele.Type() == unlockConditionType {
			*f = append((*f)[:i], (*f)[i+1:]...)
			return true
		}
	}

	return false
}

// Sort sorts the UnlockConditions in place by type.
func (f UnlockConditions[T]) Sort() {
	sort.Slice(f, func(i, j int) bool { return f[i].Type() < f[j].Type() })
}

// UnlockConditionSet is a set of UnlockCondition(s).
type UnlockConditionSet map[UnlockConditionType]UnlockCondition

// HasStorageDepositReturnCondition tells whether this set has a StorageDepositReturnUnlockCondition.
func (f UnlockConditionSet) HasStorageDepositReturnCondition() bool {
	return f.StorageDepositReturn() != nil
}

// HasExpirationCondition tells whether this set has an ExpirationUnlockCondition.
func (f UnlockConditionSet) HasExpirationCondition() bool {
	return f.Expiration() != nil
}

// CheckExpirationCondition returns the expiration return address in case an expiration condition was set and
// the future bound slot is greater than the expiration slot.
// In case the past bound slot is smaller than the expiration slot, "nil" is returned to indicate that the original owner can unlock the output.
// The range in between is not unlockable by anyone and an "ErrExpirationConditionUnlockFailed" error will be returned.
func (f UnlockConditionSet) CheckExpirationCondition(futureBoundedSlotIndex, pastBoundedSlotIndex SlotIndex) (Address, error) {
	if f.HasExpirationCondition() {
		if ok, returnAddr := f.ReturnAddressCanUnlock(futureBoundedSlotIndex); ok {
			return returnAddr, nil
		}

		if !f.OwnerAddressCanUnlock(pastBoundedSlotIndex) {
			return nil, ErrExpirationNotUnlockable
		}
	}

	//nolint:nilnil // nil, nil is ok in this context, even if it is not go idiomatic
	return nil, nil
}

// HasTimelockCondition tells whether this set has a TimelockUnlockCondition.
func (f UnlockConditionSet) HasTimelockCondition() bool {
	return f.Timelock() != nil
}

// HasManalockCondition tells whether the set has both an account address unlock
// and a timelock that is still locked at slot index.
func (f UnlockConditionSet) HasManalockCondition(accountID AccountID, slot SlotIndex) bool {
	if !f.HasTimelockUntil(slot) {
		return false
	}
	unlockAddress := f.Address()
	if unlockAddress == nil {
		return false
	}
	if unlockAddress.Address.Type() != AddressAccount {
		return false
	}
	if !unlockAddress.Address.Equal(accountID.ToAddress()) {
		return false
	}

	return true
}

// HasTimelockUntil tells us whether the set has a timelock that is still locked at slot.
func (f UnlockConditionSet) HasTimelockUntil(slot SlotIndex) bool {
	// TODO: Test this.
	timelock := f.Timelock()
	return timelock != nil && slot < timelock.Slot
}

// tells whether the given address can unlock an output containing this set of UnlockCondition(s)
// returns booleans indicating whether the given address can unlock and whether the return address can unlock.
func (f UnlockConditionSet) unlockableBy(addr Address, owner Address, pastBoundedSlot SlotIndex, futureBoundedSlot SlotIndex) bool {
	if err := f.TimelocksExpired(futureBoundedSlot); err != nil {
		return false
	}

	// if the return addr can unlock, then addr must be the return address
	if returnAddrCanUnlock, returnAddr := f.ReturnAddressCanUnlock(futureBoundedSlot); returnAddrCanUnlock {
		return addr.Equal(returnAddr)
	}

	// if the past bounded index is less than the expiration slot index, then owner can unlock
	if f.OwnerAddressCanUnlock(pastBoundedSlot) {
		return addr.Equal(owner)
	}

	return false
}

// OwnerAddressCanUnlock tells whether the target address defined in an expiration unlock condition within this set is
// allowed to unlock an Output containing this UnlockConditionSet given the past bounded slot index of the tx defined as
// the slot index of the commitment input plus the max committable age.
func (f UnlockConditionSet) OwnerAddressCanUnlock(pastBoundedSlot SlotIndex) bool {
	expUnlockCond := f.Expiration()

	// if there is not expiration unlock, then the owner can unlock.
	if expUnlockCond == nil {
		return true
	}

	if pastBoundedSlot < expUnlockCond.Slot {
		return true
	}

	return false
}

// ReturnAddressCanUnlock tells whether a sender defined in an expiration unlock condition within this set is the actual
// address which could unlock an Output containing this UnlockConditionSet given the future bounded slot index of the tx
// defined as the slot index of the commitment input plus the min committable age.
func (f UnlockConditionSet) ReturnAddressCanUnlock(futureBoundedSlotIndex SlotIndex) (bool, Address) {
	expUnlockCond := f.Expiration()

	if expUnlockCond == nil {
		return false, nil
	}

	if futureBoundedSlotIndex >= expUnlockCond.Slot {
		return true, expUnlockCond.ReturnAddress
	}

	return false, nil
}

// TimelocksExpired tells whether UnlockCondition(s) in this set which impose a timelock are expired
// in relation to the given future bounded slot index. The provided slot index is the slot index of the commitment
// input which is being spent by the transaction plus the min committable age.
func (f UnlockConditionSet) TimelocksExpired(futureBoundedSlot SlotIndex) error {
	timelock := f.Timelock()

	if timelock == nil {
		return nil
	}

	if futureBoundedSlot < timelock.Slot {
		return ierrors.WithMessagef(ErrTimelockNotExpired, "timelock expires in slot %d but the future bounded slot is only %d", timelock.Slot, futureBoundedSlot)
	}

	return nil
}

// StorageDepositReturn returns the StorageDepositReturnUnlockCondition in the set or nil.
func (f UnlockConditionSet) StorageDepositReturn() *StorageDepositReturnUnlockCondition {
	b, has := f[UnlockConditionStorageDepositReturn]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is a StorageDepositReturnUnlockCondition
	return b.(*StorageDepositReturnUnlockCondition)
}

// Address returns the AddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) Address() *AddressUnlockCondition {
	b, has := f[UnlockConditionAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an AddressUnlockCondition
	return b.(*AddressUnlockCondition)
}

// ImmutableAccount returns the ImmutableAccountUnlockCondition in the set or nil.
func (f UnlockConditionSet) ImmutableAccount() *ImmutableAccountUnlockCondition {
	b, has := f[UnlockConditionImmutableAccount]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an ImmutableAccountUnlockCondition
	return b.(*ImmutableAccountUnlockCondition)
}

// GovernorAddress returns the GovernorAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) GovernorAddress() *GovernorAddressUnlockCondition {
	b, has := f[UnlockConditionGovernorAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an GovernorAddressUnlockCondition
	return b.(*GovernorAddressUnlockCondition)
}

// StateControllerAddress returns the StateControllerAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) StateControllerAddress() *StateControllerAddressUnlockCondition {
	b, has := f[UnlockConditionStateControllerAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an StateControllerAddressUnlockCondition
	return b.(*StateControllerAddressUnlockCondition)
}

// Timelock returns the TimelockUnlockCondition in the set or nil.
func (f UnlockConditionSet) Timelock() *TimelockUnlockCondition {
	b, has := f[UnlockConditionTimelock]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an TimelockUnlockCondition
	return b.(*TimelockUnlockCondition)
}

// Expiration returns the ExpirationUnlockCondition in the set or nil.
func (f UnlockConditionSet) Expiration() *ExpirationUnlockCondition {
	b, has := f[UnlockConditionExpiration]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an ExpirationUnlockCondition
	return b.(*ExpirationUnlockCondition)
}
