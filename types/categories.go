package types

// Category represents what type of thing a release is
type Category int

// Category Constants, just for use when matching and setting categories.  DB
// has it's own representation.  The INT values here and in the DB IDs must be
// kept in sync
const (
	Unknown Category = -1
	Other   Category = 0

	Console Category = 1000
	Movies  Category = 2000
	Audio   Category = 3000
	PC      Category = 4000
	TV      Category = 5000
	XXX     Category = 6000
	Books   Category = 7000

	Other_Misc   Category = 10
	Other_Hashed Category = 20

	Console_NDS        Category = 1010
	Console_PSP        Category = 1020
	Console_Wii        Category = 1030
	Console_Xbox       Category = 1040
	Console_Xbox360    Category = 1050
	Console_WiiWareVC  Category = 1060
	Console_XBOX360DLC Category = 1070
	Console_PS3        Category = 1080
	Console_Other      Category = 1090
	Console_3DS        Category = 1110
	Console_PSVita     Category = 1120
	Console_WiiU       Category = 1130
	Console_XboxOne    Category = 1140
	Console_PS4        Category = 1180

	Movie_Foreign Category = 2010
	Movie_Other   Category = 2020
	Movie_SD      Category = 2030
	Movie_HD      Category = 2040
	Movie_3D      Category = 2050
	Movie_BluRay  Category = 2060
	Movie_DVD     Category = 2070
	Movie_WEBDL   Category = 2080

	Audio_MP3       Category = 3010
	Audio_Video     Category = 3020
	Audio_Audiobook Category = 3030
	Audio_Lossless  Category = 3040
	Audio_Other     Category = 3050
	Audio_Foreign   Category = 3060

	PC_0day         Category = 4010
	PC_ISO          Category = 4020
	PC_Mac          Category = 4030
	PC_PhoneOther   Category = 4040
	PC_Games        Category = 4050
	PC_PhoneIOS     Category = 4060
	PC_PhoneAndroid Category = 4070

	TV_WEBDL       Category = 5010
	TV_Foreign     Category = 5020
	TV_SD          Category = 5030
	TV_HD          Category = 5040
	TV_Other       Category = 5050
	TV_Sport       Category = 5060
	TV_Anime       Category = 5070
	TV_Documentary Category = 5080

	XXX_DVD      Category = 6010
	XXX_WMV      Category = 6020
	XXX_XviD     Category = 6030
	XXX_x264     Category = 6040
	XXX_Other    Category = 6050
	XXX_Imageset Category = 6060
	XXX_Packs    Category = 6070
	XXX_SD       Category = 6080
	XXX_WEBDL    Category = 6090

	Book_Ebook     Category = 7010
	Book_Comics    Category = 7020
	Book_Magazines Category = 7030
	Book_Technical Category = 7040
	Book_Other     Category = 7050
	Book_Foreign   Category = 7060
)

var categoryMap = map[Category]string{
	0:    "Other",
	1000: "Console",
	2000: "Movies",
	3000: "Audio",
	4000: "PC",
	5000: "TV",
	6000: "XXX",
	7000: "Books",
	10:   "Other_Misc",
	20:   "Other_Hashed",
	1010: "Console_NDS",
	1020: "Console_PSP",
	1030: "Console_Wii",
	1040: "Console_Xbox",
	1050: "Console_Xbox360",
	1060: "Console_WiiWareVC",
	1070: "Console_XBOX360DLC",
	1080: "Console_PS3",
	1090: "Console_Other",
	1110: "Console_3DS",
	1120: "Console_PSVita",
	1130: "Console_WiiU",
	1140: "Console_XboxOne",
	1180: "Console_PS4",
	2010: "Movie_Foreign",
	2020: "Movie_Other",
	2030: "Movie_SD",
	2040: "Movie_HD",
	2050: "Movie_3D",
	2060: "Movie_BluRay",
	2070: "Movie_DVD",
	2080: "Movie_WEBDL",
	3010: "Audio_MP3",
	3020: "Audio_Video",
	3030: "Audio_Audiobook",
	3040: "Audio_Lossless",
	3050: "Audio_Other",
	3060: "Audio_Foreign",
	4010: "PC_0day",
	4020: "PC_ISO",
	4030: "PC_Mac",
	4040: "PC_PhoneOther",
	4050: "PC_Games",
	4060: "PC_PhoneIOS",
	4070: "PC_PhoneAndroid",
	5010: "TV_WEBDL",
	5020: "TV_Foreign",
	5030: "TV_SD",
	5040: "TV_HD",
	5050: "TV_Other",
	5060: "TV_Sport",
	5070: "TV_Anime",
	5080: "TV_Documentary",
	6010: "XXX_DVD",
	6020: "XXX_WMV",
	6030: "XXX_XviD",
	6040: "XXX_x264",
	6050: "XXX_Other",
	6060: "XXX_Imageset",
	6070: "XXX_Packs",
	6080: "XXX_SD",
	6090: "XXX_WEBDL",
	7010: "Book_Ebook",
	7020: "Book_Comics",
	7030: "Book_Magazines",
	7040: "Book_Technical",
	7050: "Book_Other",
	7060: "Book_Foreign",
}

func (c Category) String() string {
	if str, ok := categoryMap[c]; ok {
		return str
	}
	return "Unknown"
}

// CategoryFromInt returns a Category corresponding to the given int or Unknown
// if there isn't one.
func CategoryFromInt(i int64) Category {
	if _, ok := categoryMap[Category(i)]; ok {
		return Category(i)
	}
	return Unknown
}
