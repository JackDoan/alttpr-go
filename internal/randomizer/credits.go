package randomizer

import (
	"github.com/JackDoan/alttpr-go/internal/helpers"
	"github.com/JackDoan/alttpr-go/internal/world"
)

// randomizeCredits picks random credit-scene texts. Mirrors PHP Randomizer::randomizeCredits.
func (r *Randomizer) randomizeCredits(w *world.World) error {
	pickFirst := func(pool []string) (string, error) {
		shuffled, err := helpers.FyShuffle(pool)
		if err != nil {
			return "", err
		}
		if len(shuffled) == 0 {
			return "", nil
		}
		return shuffled[0], nil
	}

	if s, err := pickFirst([]string{
		"the return of the king",
		"fellowship of the ring",
		"the two towers",
	}); err == nil {
		w.SetCredit("castle", s)
	}

	if s, err := pickFirst([]string{
		"the loyal priest",
		"read a book",
		"sits in own pew",
		"heal plz",
	}); err == nil {
		w.SetCredit("sanctuary", s)
	}

	name, err := pickFirst([]string{
		"sahasralah", "sabotaging", "sacahuista", "sacahuiste", "saccharase", "saccharide", "saccharify",
		"saccharine", "saccharins", "sacerdotal", "sackcloths", "salmonella", "saltarelli", "saltarello",
		"saltations", "saltbushes", "saltcellar", "saltshaker", "salubrious", "sandgrouse", "sandlotter",
		"sandstorms", "sandwiched", "sauerkraut", "schipperke", "schismatic", "schizocarp", "schmalzier",
		"schmeering", "schmoosing", "shibboleth", "shovelnose", "sahananana", "sarararara", "salamander",
		"sharshalah", "shahabadoo", "sassafrass", "saddlebags", "sandalwood", "shagadelic", "sandcastle",
		"saltpeters", "shabbiness", "shlrshlrsh", "sassyralph", "sallyacorn", "sahasrahbot", "sasharalla",
	})
	if err == nil {
		w.SetCredit("kakariko", name+"'s homecoming")
	}

	if s, err := pickFirst([]string{
		"twin lumberjacks", "fresh flapjacks", "two woodchoppers",
		"double lumberman", "lumberclones", "woodfellas", "dos axes",
	}); err == nil {
		w.SetCredit("lumberjacks", s)
	}

	smithy, err := helpers.GetRandomInt(0, 1)
	if err != nil {
		return err
	}
	if smithy == 1 {
		w.SetCredit("smithy", "the dwarven breadsmiths")
	}

	if s, err := pickFirst([]string{
		"the lost old man", "gary the old man", "Your ad here",
	}); err == nil {
		w.SetCredit("bridge", s)
	}

	if s, err := pickFirst([]string{
		"the forest thief", "dancing pickles", "flying vultures",
	}); err == nil {
		w.SetCredit("woods", s)
	}

	if s, err := pickFirst([]string{
		"venus. queen of faeries",
		"Venus was her name",
		"I'm your Venus",
		"Yeah, baby, she's got it",
		"Venus, I'm your fire",
		"Venus, At your desire",
		"Venus Love Chain",
		"Venus Crescent Beam",
	}); err == nil {
		w.SetCredit("well", s)
	}
	return nil
}
