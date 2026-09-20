package main

// PinSeed describes one seed pin: the image file it uploads (relative to
// the seed/img directory) and the title/description stored in Postgres.
type PinSeed struct {
	Slug        string
	Filename    string
	Title       string
	Description string
}

var pinSeeds = []PinSeed{
	{
		Slug:        "contrasting-volcano",
		Filename:    "pexels-clive-kim-2523249-6307488.jpg",
		Title:       "Contrasting Volcano",
		Description: "A dramatic volcano eruption in Guatemala beneath a star-filled night sky, showcasing nature's power.",
	},
	{
		Slug:        "mist-valley",
		Filename:    "pexels-eberhardgross-12365962.jpg",
		Title:       "Mist valley",
		Description: "A wide valley landscape covered in mist.",
	},
	{
		Slug:        "lake-braies",
		Filename:    "pexels-francesco-ungaro-1525041.jpg",
		Title:       "Lake Braies",
		Description: "A breathtaking view of Lake Braies with mountain reflection in the Italian Alps. Ideal for nature lovers.",
	},
	{
		Slug:        "aerial-photo",
		Filename:    "pexels-leonhard-niederwimmer-2156971331-38754765.jpg",
		Title:       "Aerial photo",
		Description: "Aerial view of a picturesque hilly landscape in Upper Austria at sunset with glowing fields and forests.",
	},
	{
		Slug:        "turquoise-sea-foam",
		Filename:    "pexels-cottonbro-9860899.jpg",
		Title:       "Turquoise Sea Foam",
		Description: "Close-up of frothy waves in brilliant turquoise water.",
	},
	{
		Slug:        "golden-desert-dunes",
		Filename:    "pexels-francesco-ungaro-17288260.jpg",
		Title:       "Stalactite Cavern",
		Description: "Rippling limestone formations lit in warm and cool tones deep underground.",
	},
	{
		Slug:        "stalactite-cavern",
		Filename:    "pexels-francesco-ungaro-998653.jpg",
		Title:       "Stalactite cavern",
		Description: "Sweeping sand dunes glowing gold under a clear desert sky.",
	},
	{
		Slug:        "joker-in-the-deck",
		Filename:    "pexels-khaled-elhowary-436473447-15318877.jpg",
		Title:       "Joker in the Deck",
		Description: "A scattered deck of playing cards with the joker on top.",
	},
	{
		Slug:        "adorable-sea",
		Filename:    "pexels-kienvirak-5213752.jpg",
		Title:       "Adorable sea",
		Description: "A figure caught mid-motion inside swirling blue light trails.",
	},
	{
		Slug:        "lost-in-the-cave",
		Filename:    "pexels-mark-direen-622749-34896823.jpg",
		Title:       "Lost in the Cave",
		Description: "A lone figure dwarfed by towering rock walls in a dim canyon cave.",
	},
	{
		Slug:        "fire-sky-over-the-shore",
		Filename:    "pexels-robert-clark-504241532-21031988.jpg",
		Title:       "Fire Sky Over the Shore",
		Description: "A dramatic sunset paints the clouds and surf in orange and violet.",
	},
	{
		Slug:        "retro-game-cartridges",
		Filename:    "pexels-stasknop-18811772.jpg",
		Title:       "Retro Game Cartridges",
		Description: "Classic NES and Famicom cartridges side by side on a dark backdrop.",
	},
	{
		Slug:        "dreamy-mountain-blur",
		Filename:    "pexels-tcuzin-36876884.jpg",
		Title:       "Dreamy Mountain Blur",
		Description: "An abstract, motion-blurred sunset over distant mountain silhouettes.",
	},
	{
		Slug:        "cavern-walkway",
		Filename:    "pexels-the-daphne-lens-2151762624-37691534.jpg",
		Title:       "Cavern Walkway",
		Description: "Close-up of frothy waves in brilliant turquoise water.",
	},
}
