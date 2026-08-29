// Package countries 提供手机号区号选择数据（等价 AppService::getAllCountries，
// 数据与 libphonenumber 支持的常用区域一致，英文显示名）。
package countries

import "strings"

type Country struct {
	ID   string `json:"id"`   // 小写 ISO 代码（对齐 PHP strtolower）
	Name string `json:"name"` // 英文显示名
	Code int    `json:"code"` // 国际区号
}

// 全量区域表：ISO -> (英文显示名, 区号)。
var regions = map[string]struct {
	Name string
	Code int
}{
	"ac": {"Ascension Island", 247}, "ad": {"Andorra", 376}, "ae": {"United Arab Emirates", 971},
	"af": {"Afghanistan", 93}, "ag": {"Antigua & Barbuda", 1268}, "ai": {"Anguilla", 1264},
	"al": {"Albania", 355}, "am": {"Armenia", 374}, "ao": {"Angola", 244},
	"ar": {"Argentina", 54}, "as": {"American Samoa", 1684}, "at": {"Austria", 43},
	"au": {"Australia", 61}, "aw": {"Aruba", 297}, "az": {"Azerbaijan", 994},
	"ba": {"Bosnia & Herzegovina", 387}, "bb": {"Barbados", 1246}, "bd": {"Bangladesh", 880},
	"be": {"Belgium", 32}, "bf": {"Burkina Faso", 226}, "bg": {"Bulgaria", 359},
	"bh": {"Bahrain", 973}, "bi": {"Burundi", 257}, "bj": {"Benin", 229},
	"bm": {"Bermuda", 1441}, "bn": {"Brunei Darussalam", 673}, "bo": {"Bolivia", 591},
	"br": {"Brazil", 55}, "bs": {"Bahamas", 1242}, "bt": {"Bhutan", 975},
	"bw": {"Botswana", 267}, "by": {"Belarus", 375}, "bz": {"Belize", 501},
	"ca": {"Canada", 1}, "cd": {"Congo (Kinshasa)", 243}, "cf": {"Central African Republic", 236},
	"cg": {"Congo (Brazzaville)", 242}, "ch": {"Switzerland", 41}, "ci": {"Côte d'Ivoire", 225},
	"ck": {"Cook Islands", 682}, "cl": {"Chile", 56}, "cm": {"Cameroon", 237},
	"cn": {"China", 86}, "co": {"Colombia", 57}, "cr": {"Costa Rica", 506},
	"cu": {"Cuba", 53}, "cv": {"Cabo Verde", 238}, "cw": {"Curaçao", 599},
	"cy": {"Cyprus", 357}, "cz": {"Czechia", 420}, "de": {"Germany", 49},
	"dj": {"Djibouti", 253}, "dk": {"Denmark", 45}, "dm": {"Dominica", 1767},
	"do": {"Dominican Republic", 1809}, "dz": {"Algeria", 213}, "ec": {"Ecuador", 593},
	"ee": {"Estonia", 372}, "eg": {"Egypt", 20}, "er": {"Eritrea", 291},
	"es": {"Spain", 34}, "et": {"Ethiopia", 251}, "fi": {"Finland", 358},
	"fj": {"Fiji", 679}, "fm": {"Micronesia", 691}, "fo": {"Faroe Islands", 298},
	"fr": {"France", 33}, "ga": {"Gabon", 241}, "gb": {"United Kingdom", 44},
	"gd": {"Grenada", 1473}, "ge": {"Georgia", 995}, "gf": {"French Guiana", 594},
	"gg": {"Guernsey", 44}, "gh": {"Ghana", 233}, "gi": {"Gibraltar", 350},
	"gl": {"Greenland", 299}, "gm": {"Gambia", 220}, "gn": {"Guinea", 224},
	"gp": {"Guadeloupe", 590}, "gq": {"Equatorial Guinea", 240}, "gr": {"Greece", 30},
	"gt": {"Guatemala", 502}, "gu": {"Guam", 1671}, "gw": {"Guinea-Bissau", 245},
	"gy": {"Guyana", 592}, "hk": {"Hong Kong", 852}, "hn": {"Honduras", 504},
	"hr": {"Croatia", 385}, "ht": {"Haiti", 509}, "hu": {"Hungary", 36},
	"id": {"Indonesia", 62}, "ie": {"Ireland", 353}, "il": {"Israel", 972},
	"in": {"India", 91}, "io": {"British Indian Ocean Territory", 246}, "iq": {"Iraq", 964},
	"ir": {"Iran", 98}, "is": {"Iceland", 354}, "it": {"Italy", 39},
	"je": {"Jersey", 44}, "jm": {"Jamaica", 1876}, "jo": {"Jordan", 962},
	"jp": {"Japan", 81}, "ke": {"Kenya", 254}, "kg": {"Kyrgyzstan", 996},
	"kh": {"Cambodia", 855}, "ki": {"Kiribati", 686}, "km": {"Comoros", 269},
	"kn": {"St Kitts & Nevis", 1869}, "kp": {"North Korea", 850}, "kr": {"South Korea", 82},
	"kw": {"Kuwait", 965}, "ky": {"Cayman Islands", 1345}, "kz": {"Kazakhstan", 7},
	"la": {"Laos", 856}, "lb": {"Lebanon", 961}, "lc": {"St Lucia", 1758},
	"li": {"Liechtenstein", 423}, "lk": {"Sri Lanka", 94}, "lr": {"Liberia", 231},
	"ls": {"Lesotho", 266}, "lt": {"Lithuania", 370}, "lu": {"Luxembourg", 352},
	"lv": {"Latvia", 371}, "ly": {"Libya", 218}, "ma": {"Morocco", 212},
	"mc": {"Monaco", 377}, "md": {"Moldova", 373}, "me": {"Montenegro", 382},
	"mg": {"Madagascar", 261}, "mh": {"Marshall Islands", 692}, "mk": {"North Macedonia", 389},
	"ml": {"Mali", 223}, "mm": {"Myanmar", 95}, "mn": {"Mongolia", 976},
	"mo": {"Macao", 853}, "mq": {"Martinique", 596}, "mr": {"Mauritania", 222},
	"ms": {"Montserrat", 1664}, "mt": {"Malta", 356}, "mu": {"Mauritius", 230},
	"mv": {"Maldives", 960}, "mw": {"Malawi", 265}, "mx": {"Mexico", 52},
	"my": {"Malaysia", 60}, "mz": {"Mozambique", 258}, "na": {"Namibia", 264},
	"nc": {"New Caledonia", 687}, "ne": {"Niger", 227}, "ng": {"Nigeria", 234},
	"ni": {"Nicaragua", 505}, "nl": {"Netherlands", 31}, "no": {"Norway", 47},
	"np": {"Nepal", 977}, "nz": {"New Zealand", 64}, "om": {"Oman", 968},
	"pa": {"Panama", 507}, "pe": {"Peru", 51}, "pf": {"French Polynesia", 689},
	"pg": {"Papua New Guinea", 675}, "ph": {"Philippines", 63}, "pk": {"Pakistan", 92},
	"pl": {"Poland", 48}, "pm": {"St Pierre & Miquelon", 508}, "pr": {"Puerto Rico", 1787},
	"ps": {"Palestine", 970}, "pt": {"Portugal", 351}, "pw": {"Palau", 680},
	"py": {"Paraguay", 595}, "qa": {"Qatar", 974}, "re": {"Réunion", 262},
	"ro": {"Romania", 40}, "rs": {"Serbia", 381}, "ru": {"Russia", 7},
	"rw": {"Rwanda", 250}, "sa": {"Saudi Arabia", 966}, "sb": {"Solomon Islands", 677},
	"sc": {"Seychelles", 248}, "sd": {"Sudan", 249}, "se": {"Sweden", 46},
	"sg": {"Singapore", 65}, "si": {"Slovenia", 386}, "sk": {"Slovakia", 421},
	"sl": {"Sierra Leone", 232}, "sm": {"San Marino", 378}, "sn": {"Senegal", 221},
	"so": {"Somalia", 252}, "sr": {"Suriname", 597}, "ss": {"South Sudan", 211},
	"st": {"São Tomé & Príncipe", 239}, "sv": {"El Salvador", 503}, "sx": {"Sint Maarten", 1721},
	"sy": {"Syria", 963}, "sz": {"Eswatini", 268}, "tc": {"Turks & Caicos Islands", 1649},
	"td": {"Chad", 235}, "tg": {"Togo", 228}, "th": {"Thailand", 66},
	"tj": {"Tajikistan", 992}, "tl": {"Timor-Leste", 670}, "tm": {"Turkmenistan", 993},
	"tn": {"Tunisia", 216}, "to": {"Tonga", 676}, "tr": {"Türkiye", 90},
	"tt": {"Trinidad & Tobago", 1868}, "tw": {"Taiwan", 886}, "tz": {"Tanzania", 255},
	"ua": {"Ukraine", 380}, "ug": {"Uganda", 256}, "us": {"United States", 1},
	"uy": {"Uruguay", 598}, "uz": {"Uzbekistan", 998}, "va": {"Vatican City", 379},
	"vc": {"St Vincent & Grenadines", 1784}, "ve": {"Venezuela", 58}, "vg": {"British Virgin Islands", 1284},
	"vi": {"US Virgin Islands", 1340}, "vn": {"Vietnam", 84}, "vu": {"Vanuatu", 678},
	"wf": {"Wallis & Futuna", 681}, "ws": {"Samoa", 685}, "ye": {"Yemen", 967},
	"za": {"South Africa", 27}, "zm": {"Zambia", 260}, "zw": {"Zimbabwe", 263},
}

// All 返回按 ISO 代码排序的全量国家列表。
func All() []Country {
	out := make([]Country, 0, len(regions))
	for id, r := range regions {
		out = append(out, Country{ID: id, Name: r.Name, Code: r.Code})
	}
	// 插入排序（数据量小且初始化一次）
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].ID < out[j-1].ID; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// Codes 返回 id -> name 映射（等价 getAllCountryCodes）。
func Codes() map[string]string {
	m := make(map[string]string, len(regions))
	for id, r := range regions {
		m[id] = r.Name
	}
	return m
}

// IsValidCountryCode 判断国家代码是否有效。
func IsValidCountryCode(id string) bool {
	_, ok := regions[strings.ToLower(id)]
	return ok
}
