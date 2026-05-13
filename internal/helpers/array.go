package helpers

// MtShuffle returns a new slice with elements drawn from `s` in random
// order using the VT-style algorithm (pick-and-remove). Mirrors
// app/Helpers/array.php:mt_shuffle.
func MtShuffle[T any](s []T) ([]T, error) {
	src := append([]T(nil), s...)
	out := make([]T, 0, len(s))
	for len(src) > 0 {
		idx, err := GetRandomInt(0, len(src)-1)
		if err != nil {
			return nil, err
		}
		out = append(out, src[idx])
		src = append(src[:idx], src[idx+1:]...)
	}
	return out, nil
}

// FyShuffle returns a new slice shuffled with Fisher-Yates. Mirrors
// app/Helpers/array.php:fy_shuffle.
func FyShuffle[T any](s []T) ([]T, error) {
	out := append([]T(nil), s...)
	for i := len(out) - 1; i >= 0; i-- {
		r, err := GetRandomInt(0, i)
		if err != nil {
			return nil, err
		}
		out[i], out[r] = out[r], out[i]
	}
	return out, nil
}

// WeightedRandomPick selects `pick` items from `items` weighted by `weights`.
// Mirrors app/Helpers/array.php:weighted_random_pick.
func WeightedRandomPick[T any](items []T, weights []int, pick int) ([]T, error) {
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	out := make([]T, 0, pick)
	for range pick {
		r, err := GetRandomInt(1, totalWeight)
		if err != nil {
			return nil, err
		}
		cur := 0
		for j, w := range weights {
			cur += w
			if r <= cur {
				out = append(out, items[j])
				break
			}
		}
	}
	return out, nil
}
