package rooms

// Hushed reports whether a one-off moment holds the room still, so the
// background patter stops (DESIGN §7.2): ZAP, BYTE, the cage falling, a dog
// fight, or the room having ended.
func Hushed(r Room) bool {
	b := r.Shared()
	if b.Out != Playing || b.Freddy.Gone {
		return true
	}
	switch r := r.(type) {
	case *FishbowlRoom:
		return r.Phase == Zapped
	case *BookshelfRoom:
		return r.Bite > 0
	case *BirdcageRoom:
		return r.Cage.Falling && !r.Cage.Broken && !r.Dog.Out()
	}
	return false
}
