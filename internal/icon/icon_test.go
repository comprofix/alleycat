package icon

import "testing"

func TestImages(t *testing.T) {
	imgs, err := Images()
	if err != nil {
		t.Fatal(err)
	}
	want := []int{16, 32, 48, 64, 128, 256}
	if len(imgs) != len(want) {
		t.Fatalf("got %d images, want %d", len(imgs), len(want))
	}
	for i, img := range imgs {
		b := img.Bounds()
		if b.Dx() != want[i] || b.Dy() != want[i] {
			t.Errorf("image %d is %v, want %d×%d", i, b.Size(), want[i], want[i])
		}
	}
	// The 48 is the 16 tripled: each 3×3 block is one colour.
	small, big := imgs[0], imgs[2]
	for y := range 16 {
		for x := range 16 {
			if small.At(x, y) != big.At(x*3+2, y*3+2) {
				t.Fatalf("48 px icon at (%d,%d) doesn't match the 16 px one", x*3+2, y*3+2)
			}
		}
	}
}
