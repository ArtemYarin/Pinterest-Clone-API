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
		Slug:        "lost-in-the-cave",
		Filename:    "pexels-cottonbro-9860899.jpg",
		Title:       "Lost in the cave",
		Description: "An ambient photo of travaller in the cave.",
	},
	{
		Slug:        "stalactite-cavern",
		Filename:    "pexels-francesco-ungaro-17288260.jpg",
		Title:       "Stalactite Cavern",
		Description: "Rippling limestone formations lit in warm and cool tones deep underground.",
	},
	{
		Slug:        "golden-dunes",
		Filename:    "pexels-francesco-ungaro-998653.jpg",
		Title:       "Golden dunes",
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
		Description: "Just sea.",
	},
	{
		Slug:        "catacombs",
		Filename:    "pexels-mark-direen-622749-34896823.jpg",
		Title:       "Catacombs",
		Description: "An atmospheric photo of catacombs.",
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
		Slug:        "neon-miniature",
		Filename:    "pexels-tcuzin-36876884.jpg",
		Title:       "Neon miniature",
		Description: "High quality miniature with neon painting",
	},
	{
		Slug:        "water",
		Filename:    "pexels-the-daphne-lens-2151762624-37691534.jpg",
		Title:       "Water",
		Description: "Close-up of frothy waves in brilliant turquoise water.",
	},
}
