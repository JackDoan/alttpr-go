package randomizer

import (
	"testing"

	"github.com/JackDoan/alttpr-go/internal/boss"
	"github.com/JackDoan/alttpr-go/internal/item"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// TestTriforceHunt_PoolAndWin verifies that triforce-hunt mode places the
// configured count of TriforcePieces and that the win condition fires
// correctly with N pieces collected.
func TestTriforceHunt_PoolAndWin(t *testing.T) {
	for _, goal := range []string{"triforce-hunt", "ganonhunt"} {
		t.Run(goal, func(t *testing.T) {
			ir := item.NewRegistry()
			br := boss.NewRegistry()
			opts := world.DefaultStandardOptions()
			opts.Goal = goal
			opts.TriforcePieces = 30
			opts.TriforceGoal = 20
			w := world.NewStandard(opts, ir, br)
			r, err := New([]*world.World{w}, ir, br)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if err := r.Randomize(); err != nil {
				t.Fatalf("Randomize: %v", err)
			}

			// Count placed TriforcePieces.
			placed := 0
			for _, loc := range w.Locations().All() {
				if loc.HasItem() && loc.Item().Name == "TriforcePiece" {
					placed++
				}
			}
			if placed != 30 {
				t.Errorf("%s: placed %d TriforcePieces, want 30", goal, placed)
			}

			// Verify win condition: empty inventory → false; 19 pieces → false (need 20).
			emptyInv := item.NewCollection()
			if w.WinCondition(emptyInv) {
				t.Errorf("%s: empty inventory should not win", goal)
			}

			// For triforce-hunt, having 20 pieces + reachable NE Light World should win.
			if goal == "triforce-hunt" {
				// Add 20 TriforcePieces + RescueZelda (which unblocks region access).
				tfp, _ := ir.Get("TriforcePiece", w.ID())
				rescue, _ := ir.Get("RescueZelda", w.ID())
				withPieces := item.NewCollection()
				withPieces.SetChecksForWorld(w.ID())
				withPieces.Add(rescue)
				for i := 0; i < 20; i++ {
					withPieces.Add(tfp)
				}
				if !w.WinCondition(withPieces) {
					t.Errorf("triforce-hunt: 20 pieces + RescueZelda should win")
				}

				// 19 pieces should NOT win.
				withFewer := item.NewCollection()
				withFewer.SetChecksForWorld(w.ID())
				withFewer.Add(rescue)
				for i := 0; i < 19; i++ {
					withFewer.Add(tfp)
				}
				if w.WinCondition(withFewer) {
					t.Errorf("triforce-hunt: 19 pieces should NOT win")
				}
			}

			// Defeating Ganon (the universal escape hatch) should always win.
			defeat, _ := ir.Get("DefeatGanon", w.ID())
			withDefeat := item.NewCollection(defeat)
			withDefeat.SetChecksForWorld(w.ID())
			if !w.WinCondition(withDefeat) {
				t.Errorf("%s: DefeatGanon should win", goal)
			}
		})
	}
}
