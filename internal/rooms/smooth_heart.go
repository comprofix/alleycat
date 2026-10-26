package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// heartArt is the heart room's pictures in one mood, the room's own or the
// kiss's (DESIGN §3.5, "The kiss"), each drawn once and laid many times.
type heartArt struct {
	back   *ebiten.Image
	hearts [2]*ebiten.Image // intact, broken
}

// heartInks are the colours of a heart platform and the cherubs in one
// mood: the heart's body, its lit lobe, its glint and its shade; a broken
// heart's two halves, lit and dark; the cherub's skin and shade and its
// bow.
type heartInks struct {
	body, lit, glint, shade int
	brokenLit, brokenDark   int
	skin, skinShade, bow    int
}

var (
	// roomInks are the heart room's: Cerise hearts lit Flamingo, broken
	// ones Navy lit Indigo, cherubs Coral with a Cerise bow.
	roomInks = heartInks{
		resurrect64.Cerise, resurrect64.Flamingo, resurrect64.White, resurrect64.Mulberry,
		resurrect64.Indigo, resurrect64.Navy,
		resurrect64.Coral, resurrect64.Copper, resurrect64.Cerise,
	}
	// kissInks are the kiss's, as heart_kiss and cherub_kiss recolour
	// them: Grape hearts lit Violet, broken ones Ink, cherubs Grape with an
	// Ink bow.
	kissInks = heartInks{
		resurrect64.Grape, resurrect64.Violet, resurrect64.Violet, resurrect64.Plum,
		resurrect64.Shadow, resurrect64.Ink,
		resurrect64.Grape, resurrect64.Plum, resurrect64.Ink,
	}
)

// newHeartArt draws the heart room's backdrop and hearts in inks c on
// ground: the lattice of little hearts when lattice is set, and the
// border of cherubs along the top and down both sides, where the classic
// look lays them.
func newHeartArt(c heartInks, ground int, lattice bool) *heartArt {
	a := &heartArt{back: ebiten.NewImage(gfx.ScreenWidth*entity.K, gfx.ScreenHeight*entity.K)}
	b := a.back
	b.Fill(gfx.Palette[ground])
	w, h := float64(b.Bounds().Dx()), float64(b.Bounds().Dy())
	for i := range 8 { // a little darker towards the foot of the screen
		fadeRect(b, 0, h/2+float64(i)*h/16, w, h/16, resurrect64.Ink, 0.02*float64(i+1))
	}
	if lattice {
		for y, row := 4, 0; y < gfx.ScreenHeight; y, row = y+12, row+1 {
			for x := 4 + (row%2)*12; x < gfx.ScreenWidth; x += 24 {
				shapes.FillPoly(b, resurrect64.Forest, entity.HeartPts(k(x)-1, k(y)+1, 14, 9)...)
				shapes.FillPoly(b, resurrect64.Leaf, entity.HeartPts(k(x)-1, k(y), 13, 7)...)
			}
		}
	}
	cherub := entity.Bake(cherubW, cherubH, resurrect64.Ink, func(p entity.Pen) { drawCherub(p, c) })
	for x := 0; x < gfx.ScreenWidth; x += cherubW {
		entity.DrawBaked(b, cherub, x, 0, x >= gfx.ScreenWidth/2)
	}
	for y := cherubSideY; y < gfx.ScreenHeight; y += cherubH {
		entity.DrawBaked(b, cherub, 0, y, false)
		entity.DrawBaked(b, cherub, gfx.ScreenWidth-cherubW, y, true)
	}
	cherub.Deallocate()
	a.hearts[0] = entity.Bake(config.HeartW, config.HeartH, entity.NoOutline, func(p entity.Pen) { drawHeartPlatform(p, c) })
	a.hearts[1] = entity.Bake(config.HeartW, config.HeartH, entity.NoOutline, func(p entity.Pen) { drawBrokenHeart(p, c) })
	return a
}

// drawHeartPlatform draws an intact heart platform in a 64×32 box, its
// lobes along the top: shaded down its right side, lit on its left lobe
// with a glint.
func drawHeartPlatform(p entity.Pen, c heartInks) {
	p.Poly(c.shade, entity.HeartPts(0, 0, 56, 31)...)
	p.Poly(c.body, entity.HeartPts(0, 0, 52, 28)...)
	p.Ellipse(14, 8, 9, 4.5, -0.5, c.lit)
	p.Circle(10, 6, 2.2, c.glint)
}

// drawBrokenHeart draws a broken heart platform in a 64×32 box: the two
// halves split down a jagged crack and pulled apart.
func drawBrokenHeart(p entity.Pen, c heartInks) {
	pts := entity.HeartPts(0, 0, 56, 31)
	crack := []float64{28, 8, 25, 13, 30, 18, 26, 23, 28, 31}
	n := len(pts) / 2
	var right, left []float64
	for i := range n / 2 { // the right side, from the dip to the point
		right = append(right, pts[2*i]+3, pts[2*i+1])
	}
	for i := n / 2; i < n; i++ { // the left side, from the point to the dip
		left = append(left, pts[2*i]-3, pts[2*i+1])
	}
	for i := len(crack) - 2; i >= 0; i -= 2 {
		right = append(right, crack[i]+3, crack[i+1])
	}
	for i := 0; i < len(crack); i += 2 {
		left = append(left, crack[i]-3, crack[i+1])
	}
	p.Poly(c.brokenDark, left...)
	p.Poly(c.brokenDark, right...)
	p.Ellipse(12, 8, 7, 3.5, -0.5, c.brokenLit)
	p.Ellipse(38, 7, 5, 2.5, 0.4, c.brokenLit)
}

// drawCherub draws a cherub with its bow, facing right in a 128×96 box,
// in inks c: white wings, a chubby body under curly hair, an arrow drawn
// on the bow.
func drawCherub(p entity.Pen, c heartInks) {
	// The wing: three rows of feathers fanned out behind, the lowest
	// shortest, each shaded along its lower edge.
	for _, f := range [][4]float64{{26, 20, 17, -0.35}, {25, 30, 17, 0}, {28, 40, 13, 0.35}} {
		p.Ellipse(f[0], f[1]+1.5, f[2], 6, f[3], resurrect64.Fog)
		p.Ellipse(f[0], f[1], f[2], 5.5, f[3], resurrect64.White)
	}
	// The legs, then the body and the arm out to the bow.
	for _, x := range []float64{42, 62} {
		p.Ellipse(x, 55, 6.5, 4.5, 0, c.skinShade)
		p.Ellipse(x-1, 54, 5, 3.5, 0, c.skin)
	}
	p.Ellipse(54, 40, 15, 14, 0, c.skinShade)
	p.Ellipse(51, 38, 13, 12, 0, c.skin)
	p.Stroke(c.skinShade, 7, false, 58, 34, 84, 34)
	p.Stroke(c.skin, 5, false, 58, 33, 84, 33)
	// The head and its curls.
	p.Circle(57, 15, 11, c.skinShade)
	p.Circle(56, 14, 10, c.skin)
	for _, h := range [][3]float64{{47, 8, 5}, {54, 4, 5.5}, {62, 5, 5}, {67, 10, 4}} {
		p.Circle(h[0], h[1], h[2], resurrect64.Honey)
	}
	p.Circle(57, 3, 2, resurrect64.Caramel)
	p.Circle(52, 14, 1.6, resurrect64.Ink)
	p.Circle(61, 14, 1.6, resurrect64.Ink)
	p.Ellipse(49, 19, 2.5, 1.6, 0, resurrect64.Flamingo)
	p.Ellipse(64, 19, 2.5, 1.6, 0, resurrect64.Flamingo)
	// The bow, its string drawn back to the hand, and the arrow.
	p.Stroke(c.bow, 4, true, 96, 6, 104, 18, 107, 34, 104, 50, 96, 60)
	p.Stroke(resurrect64.Fog, 1.5, false, 96, 6, 86, 33, 96, 60)
	p.Stroke(resurrect64.Ink, 2.5, false, 80, 33, 116, 33)
	p.Poly(c.bow, 114, 28, 125, 33, 114, 38)
	p.Stroke(resurrect64.White, 2, false, 80, 33, 76, 28)
	p.Stroke(resurrect64.White, 2, false, 80, 33, 76, 38)
}

// drawGift draws a gift box in a 96×60 box: a red box shaded down its
// right side, its lid, and a honey ribbon tied in a bow on top.
func drawGift(p entity.Pen) {
	p.Poly(resurrect64.Cerise, 12, 20, 72, 20, 72, 59, 12, 59)
	p.Poly(resurrect64.Grape, 72, 20, 80, 20, 80, 59, 72, 59)
	p.Stroke(resurrect64.Honey, 6, false, 47, 20, 47, 59)
	p.Stroke(resurrect64.Honey, 5, false, 12, 39, 72, 39)
	p.Stroke(resurrect64.Caramel, 5, false, 72, 39, 80, 39)
	p.Poly(resurrect64.Flamingo, 8, 12, 84, 12, 84, 17, 8, 17)
	p.Poly(resurrect64.Grape, 8, 17, 84, 17, 84, 21, 8, 21)
	p.Stroke(resurrect64.Honey, 6, false, 47, 12, 47, 21)
	for _, side := range []float64{-1, 1} {
		p.Stroke(resurrect64.Honey, 3, true, 47, 11, 47+side*6, 3, 47+side*15, 2, 47+side*15, 9, 47, 11)
	}
	p.Circle(47, 11, 3.5, resurrect64.Caramel)
}

// arrowTails are where the arrow's tail is as it grows in, frame by frame,
// in its 64×32 box; its head is always at the right.
var arrowTails = [config.ArrowFrames][2]float64{{50, 22}, {44, 18}, {34, 14}, {22, 10}, {13, 6}, {3, 2}}

// drawArrow draws Cupid's arrow facing right, sloping down as it flies, in
// frame's length: an ink shaft, a red head and, once it is all in, white
// fletching.
func drawArrow(p entity.Pen, frame int) {
	t := arrowTails[frame]
	p.Stroke(resurrect64.Ink, 3, false, t[0], t[1], 56, 24)
	p.Poly(resurrect64.Cerise, 50, 19, 63, 27, 50, 30)
	if frame >= config.ArrowFrames-2 {
		for _, d := range []float64{0, 5} {
			p.Stroke(resurrect64.White, 2, false, t[0]+d+1, t[1]+d*0.4, t[0]+d-3, t[1]+d*0.4-5)
			p.Stroke(resurrect64.White, 2, false, t[0]+d+1, t[1]+d*0.4, t[0]+d+1, t[1]+d*0.4+5)
		}
	}
}

// drawHeart draws the heart room in the new look, as the classic look does
// (DESIGN §3.5): the hearts, the gifts, the cats, the arrow and Freddy,
// blue while he carries a gift.
func (r *SmoothRenderer) drawHeart(dst *ebiten.Image, h *HeartRoom) {
	if r.heart == nil {
		r.heart = newHeartArt(roomInks, gfx.HeartScheme.Wall, true)
	}
	r.drawHeartScene(dst, h, r.heart)
	for i := range h.Cats {
		co := entity.RivalCoat
		if i == config.Felicia {
			co = entity.FeliciaCoat
		}
		r.drawHeartCat(dst, &h.Cats[i], co)
	}
	if a := &h.Arrow; a.Active {
		pic := entity.Picture{X: a.X, Y: a.Y, W: config.ArrowW, H: config.ArrowH, Flip: a.Dir < 0, Outline: entity.NoOutline}
		r.canvas.Draw(dst, pic, func(p entity.Pen) { drawArrow(p, a.Frame()) })
	}
	f := h.Freddy
	co := entity.FreddyCoat
	if h.Carrying >= 0 {
		co = entity.GiftCoat
	}
	r.freddy.DrawIn(dst, f, co)
}

// drawHeartScene draws the backdrop, the hearts and the gifts lying on the
// rows, in art's mood.
func (r *SmoothRenderer) drawHeartScene(dst *ebiten.Image, h *HeartRoom, art *heartArt) {
	dst.DrawImage(art.back, nil)
	for row, line := range config.HeartRows {
		for col := range config.HeartCols {
			img := art.hearts[0]
			if h.Broken[row][col] {
				img = art.hearts[1]
			}
			entity.DrawBaked(dst, img, config.HeartX+col*config.HeartW, line+config.HeartDrawDY, false)
		}
	}
	if r.gift == nil {
		r.gift = entity.Bake(config.GiftW, config.GiftH, resurrect64.Ink, drawGift)
	}
	for _, g := range h.Gifts {
		if g.On {
			entity.DrawBaked(dst, r.gift, g.X, g.Y, false)
		}
	}
}

// drawHeartCat draws a heart room cat in coat co, standing on its row as
// Freddy does: walking, stopped, or leaping back once it has bumped him.
func (r *SmoothRenderer) drawHeartCat(dst *ebiten.Image, c *HeartCat, co entity.Coat) {
	if c.Away {
		return
	}
	ps := entity.CatWalking
	switch {
	case c.Bump > 0:
		ps = entity.CatLeaping
	case c.Dir == 0:
		ps = entity.CatStanding
	}
	y := c.Y + config.HeartStandDY - config.HeartCatDY
	r.freddy.DrawCat(dst, c.X, y, c.Dir < 0, ps, c.Frame, co, entity.NoOutline)
}

// DrawHeartKiss draws the heart room as it stands at the kiss, as the
// classic look does, in the kiss colours (DESIGN §3.5, "The kiss"): the
// room is frozen, Freddy and Felicia are in the big heart, and the rivals
// turn silver.
func (r *SmoothRenderer) DrawHeartKiss(dst *ebiten.Image, h *HeartRoom) {
	if r.kiss == nil {
		r.kiss = newHeartArt(kissInks, gfx.KissScheme.Wall, false)
	}
	r.drawHeartScene(dst, h, r.kiss)
	for i := range config.Felicia {
		r.drawHeartCat(dst, &h.Cats[i], entity.SilverCoat)
	}
}

// kissHearts are the kiss's hearts by their classic sprites' names, each
// box's width and height and the part of it the heart fills.
var kissHearts = map[string][4]int{
	"heart_small": {16, 8, 16, 7},
	"heart_mid":   {32, 16, 30, 14},
	"heart_big":   {56, 32, 54, 28},
}

// DrawKissHeart lays one of the kiss's hearts, by its classic sprite's
// name, with its top-left at (x, y) on the layout, and Freddy and Felicia
// in it when cats is set, where the classic look puts them.
func (r *SmoothRenderer) DrawKissHeart(dst *ebiten.Image, sprite string, x, y int, cats bool) {
	if r.kissArt == nil {
		r.kissArt = map[string]*ebiten.Image{}
		for name, s := range kissHearts {
			w, h := float64(s[2]*entity.K), float64(s[3]*entity.K)
			r.kissArt[name] = entity.Bake(s[0], s[1], entity.NoOutline, func(p entity.Pen) {
				p.Poly(resurrect64.White, entity.HeartPts(0, 0, w, h)...)
				p.Poly(resurrect64.Plum, entity.HeartPts(4, 3, w-8, h-7)...)
				p.Poly(resurrect64.Grape, entity.HeartPts(4, 3, w-12, h-11)...)
				p.Ellipse(w*0.25, h*0.22, w*0.12, h*0.08, -0.5, resurrect64.Violet)
			})
		}
		r.kissArt["cats"] = entity.Bake(32, 13, entity.NoOutline, entity.DrawKissCats)
	}
	entity.DrawBaked(dst, r.kissArt[sprite], x, y, false)
	if cats {
		entity.DrawBaked(dst, r.kissArt["cats"], x+config.KissCatsDX, y+config.KissCatsDY, false)
	}
}
