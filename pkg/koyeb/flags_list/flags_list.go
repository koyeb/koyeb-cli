package flags_list

import "github.com/sirupsen/logrus"

// Flag is an interface which represents a flag passed on the command line. T is
// the type of the item to update, e.g. koyeb.DeploymentEnv or
// koyeb.DeploymentPort.
type Flag[T any] interface {
	// Should return the flag as it was passed on the command line
	String() string
	// Should return true if the flag is a deletion flag (e.g. --env !KEY)
	IsDeletionFlag() bool
	// Compare a flag with an item of the list
	IsEqualTo(T) bool
	// Update the item with the flag
	UpdateItem(*T)
	// Create a new item from the flag
	CreateNewItem() *T
}

// ParseListFlags is a generic function which takes a list of flags and a list of existing items, and returns the updated list of items.
// It is used to parse the flags --env, --checks, --routes and --ports of `koyeb service update`.
// The function will:
// - Remove the items marked for deletion from the existingItems list
// - If the flag corresponds to an existing item, update the item from the existingItems list
// - Otherwise, create a new item and append it to the existingItems list
func ParseListFlags[ItemType any](
	flags []Flag[ItemType],
	existingItems []ItemType,
) []ItemType {

	for _, flag := range flags {
		if flag.IsDeletionFlag() {
			newItems, found := deleteFromList(flag, existingItems)
			if !found {
				logrus.Warnf("The flag \"%s\" attempts to remove an item, but this item is not configured for the service. This flag will be ignored.", flag)
			}
			existingItems = newItems
		} else {
			// Search if the flag corresponds to an existing item. If yes, update the item.
			found := false
			for idx := range existingItems {
				if flag.IsEqualTo(existingItems[idx]) {
					found = true
					flag.UpdateItem(&existingItems[idx])
					break
				}
			}
			if !found {
				existingItems = append(existingItems, *flag.CreateNewItem())
			}
		}
	}
	return existingItems
}

func deleteFromList[ItemType any](flag Flag[ItemType], existingItems []ItemType) ([]ItemType, bool) {
	for idx, item := range existingItems {
		if flag.IsEqualTo(item) {
			withoutItem := make([]ItemType, 0, len(existingItems)-1)
			withoutItem = append(withoutItem, existingItems[:idx]...)
			withoutItem = append(withoutItem, existingItems[idx+1:]...)
			return withoutItem, true
		}
	}
	return existingItems, false
}
