package regex

import (
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/pkg/validation/validators"
)

func init() { validation.Register(ASCIIRegexProvider) }

var ASCIIRegexProvider = validators.NewRegexpProvider(regexASCII, "ascii", "")

func init() { validation.Register(AlphaRegexProvider) }

var AlphaRegexProvider = validators.NewRegexpProvider(regexAlpha, "alpha", "")

func init() { validation.Register(AlphaNumericRegexProvider) }

var AlphaNumericRegexProvider = validators.NewRegexpProvider(regexAlphaNumeric, "alphaNumeric", "")

func init() { validation.Register(AlphaNumericSpaceRegexProvider) }

var AlphaNumericSpaceRegexProvider = validators.NewRegexpProvider(regexAlphaNumericSpace, "alphaNumericSpace", "")

func init() { validation.Register(AlphaSpaceRegexProvider) }

var AlphaSpaceRegexProvider = validators.NewRegexpProvider(regexAlphaSpace, "alphaSpace", "")

func init() { validation.Register(AlphaUnicodeRegexProvider) }

var AlphaUnicodeRegexProvider = validators.NewRegexpProvider(regexAlphaUnicode, "alphaUnicode", "")

func init() { validation.Register(AlphaUnicodeNumericRegexProvider) }

var AlphaUnicodeNumericRegexProvider = validators.NewRegexpProvider(regexAlphaUnicodeNumeric, "alphaUnicodeNumeric", "")

func init() { validation.Register(Base32RegexProvider) }

var Base32RegexProvider = validators.NewRegexpProvider(regexBase32, "base32", "")

func init() { validation.Register(Base64RegexProvider) }

var Base64RegexProvider = validators.NewRegexpProvider(regexBase64, "base64", "")

func init() { validation.Register(Base64RawURLRegexProvider) }

var Base64RawURLRegexProvider = validators.NewRegexpProvider(regexBase64RawURL, "base64RawUrl", "")

func init() { validation.Register(Base64URLRegexProvider) }

var Base64URLRegexProvider = validators.NewRegexpProvider(regexBase64URL, "base64Url", "")

func init() { validation.Register(Bic2014RegexProvider) }

var Bic2014RegexProvider = validators.NewRegexpProvider(regexBic2014, "bic2014", "")

func init() { validation.Register(Bic2022RegexProvider) }

var Bic2022RegexProvider = validators.NewRegexpProvider(regexBic2022, "bic2022", "")

func init() { validation.Register(BtcAddressRegexProvider) }

var BtcAddressRegexProvider = validators.NewRegexpProvider(regexBtcAddress, "btcAddress", "")

func init() { validation.Register(BtcAddressLowerBech32RegexProvider) }

var BtcAddressLowerBech32RegexProvider = validators.NewRegexpProvider(regexBtcAddressLowerBech32, "btcAddressLowerBech32", "")

func init() { validation.Register(BtcAddressUpperBech32RegexProvider) }

var BtcAddressUpperBech32RegexProvider = validators.NewRegexpProvider(regexBtcAddressUpperBech32, "btcAddressUpperBech32", "")

func init() { validation.Register(CMYKRegexProvider) }

var CMYKRegexProvider = validators.NewRegexpProvider(regexCMYK, "cmyk", "")

func init() { validation.Register(CVERegexProvider) }

var CVERegexProvider = validators.NewRegexpProvider(regexCVE, "cve", "")

func init() { validation.Register(CronRegexProvider) }

var CronRegexProvider = validators.NewRegexpProvider(regexCron, "cron", "")

func init() { validation.Register(DataURIRegexProvider) }

var DataURIRegexProvider = validators.NewRegexpProvider(regexDataURI, "dataUri", "")

func init() { validation.Register(DnsRFC1035LabelRegexProvider) }

var DnsRFC1035LabelRegexProvider = validators.NewRegexpProvider(regexDnsRFC1035Label, "dnsRfc1035Label", "")

func init() { validation.Register(E164RegexProvider) }

var E164RegexProvider = validators.NewRegexpProvider(regexE164, "e164", "")

func init() { validation.Register(EINRegexProvider) }

var EINRegexProvider = validators.NewRegexpProvider(regexEIN, "ein", "")

func init() { validation.Register(EmailRegexProvider) }

var EmailRegexProvider = validators.NewRegexpProvider(regexEmail, "email", "")

func init() { validation.Register(EthAddressRegexProvider) }

var EthAddressRegexProvider = validators.NewRegexpProvider(regexEthAddress, "ethAddress", "")

func init() { validation.Register(EthAddressLowerRegexProvider) }

var EthAddressLowerRegexProvider = validators.NewRegexpProvider(regexEthAddressLower, "ethAddressLower", "")

func init() { validation.Register(EthAddressUpperRegexProvider) }

var EthAddressUpperRegexProvider = validators.NewRegexpProvider(regexEthAddressUpper, "ethAddressUpper", "")

func init() { validation.Register(FqdnRFC1123RegexProvider) }

var FqdnRFC1123RegexProvider = validators.NewRegexpProvider(regexFqdnRFC1123, "fqdnRfc1123", "")

func init() { validation.Register(HSLRegexProvider) }

var HSLRegexProvider = validators.NewRegexpProvider(regexHSL, "hsl", "")

func init() { validation.Register(HSLARegexProvider) }

var HSLARegexProvider = validators.NewRegexpProvider(regexHSLA, "hsla", "")

func init() { validation.Register(HTMLRegexProvider) }

var HTMLRegexProvider = validators.NewRegexpProvider(regexHTML, "html", "")

func init() { validation.Register(HTMLEncodedRegexProvider) }

var HTMLEncodedRegexProvider = validators.NewRegexpProvider(regexHTMLEncoded, "htmlEncoded", "")

func init() { validation.Register(HexColorRegexProvider) }

var HexColorRegexProvider = validators.NewRegexpProvider(regexHexColor, "hexColor", "")

func init() { validation.Register(HexadecimalRegexProvider) }

var HexadecimalRegexProvider = validators.NewRegexpProvider(regexHexadecimal, "hexadecimal", "")

func init() { validation.Register(HostnameRFC1123RegexProvider) }

var HostnameRFC1123RegexProvider = validators.NewRegexpProvider(regexHostnameRFC1123, "hostnameRfc1123", "")

func init() { validation.Register(HostnameRFC952RegexProvider) }

var HostnameRFC952RegexProvider = validators.NewRegexpProvider(regexHostnameRFC952, "hostnameRfc952", "")

func init() { validation.Register(Isbn10RegexProvider) }

var Isbn10RegexProvider = validators.NewRegexpProvider(regexIsbn10, "isbn10", "")

func init() { validation.Register(Isbn13RegexProvider) }

var Isbn13RegexProvider = validators.NewRegexpProvider(regexIsbn13, "isbn13", "")

func init() { validation.Register(IssnRegexProvider) }

var IssnRegexProvider = validators.NewRegexpProvider(regexIssn, "issn", "")

func init() { validation.Register(JWTRegexProvider) }

var JWTRegexProvider = validators.NewRegexpProvider(regexJWT, "jwt", "")

func init() { validation.Register(LatitudeRegexProvider) }

var LatitudeRegexProvider = validators.NewRegexpProvider(regexLatitude, "latitude", "")

func init() { validation.Register(LongitudeRegexProvider) }

var LongitudeRegexProvider = validators.NewRegexpProvider(regexLongitude, "longitude", "")

func init() { validation.Register(MD4RegexProvider) }

var MD4RegexProvider = validators.NewRegexpProvider(regexMD4, "md4", "")

func init() { validation.Register(MD5RegexProvider) }

var MD5RegexProvider = validators.NewRegexpProvider(regexMD5, "md5", "")

func init() { validation.Register(MongodbConnStringRegexProvider) }

var MongodbConnStringRegexProvider = validators.NewRegexpProvider(regexMongodbConnString, "mongodbConnString", "")

func init() { validation.Register(MongodbIdRegexProvider) }

var MongodbIdRegexProvider = validators.NewRegexpProvider(regexMongodbId, "mongodbId", "")

func init() { validation.Register(MultibyteRegexProvider) }

var MultibyteRegexProvider = validators.NewRegexpProvider(regexMultibyte, "multibyte", "")

func init() { validation.Register(NumberRegexProvider) }

var NumberRegexProvider = validators.NewRegexpProvider(regexNumber, "number", "")

func init() { validation.Register(NumericRegexProvider) }

var NumericRegexProvider = validators.NewRegexpProvider(regexNumeric, "numeric", "")

func init() { validation.Register(PrintableASCIIRegexProvider) }

var PrintableASCIIRegexProvider = validators.NewRegexpProvider(regexPrintableASCII, "printableAscii", "")

func init() { validation.Register(RGBRegexProvider) }

var RGBRegexProvider = validators.NewRegexpProvider(regexRGB, "rgb", "")

func init() { validation.Register(RGBARegexProvider) }

var RGBARegexProvider = validators.NewRegexpProvider(regexRGBA, "rgba", "")

func init() { validation.Register(RIPEMD128RegexProvider) }

var RIPEMD128RegexProvider = validators.NewRegexpProvider(regexRIPEMD128, "ripemd128", "")

func init() { validation.Register(RIPEMD160RegexProvider) }

var RIPEMD160RegexProvider = validators.NewRegexpProvider(regexRIPEMD160, "ripemd160", "")

func init() { validation.Register(SHA256RegexProvider) }

var SHA256RegexProvider = validators.NewRegexpProvider(regexSHA256, "sha256", "")

func init() { validation.Register(SHA384RegexProvider) }

var SHA384RegexProvider = validators.NewRegexpProvider(regexSHA384, "sha384", "")

func init() { validation.Register(SHA512RegexProvider) }

var SHA512RegexProvider = validators.NewRegexpProvider(regexSHA512, "sha512", "")

func init() { validation.Register(SSNRegexProvider) }

var SSNRegexProvider = validators.NewRegexpProvider(regexSSN, "ssn", "")

func init() { validation.Register(SemverRegexProvider) }

var SemverRegexProvider = validators.NewRegexpProvider(regexSemver, "semver", "")

func init() { validation.Register(SplitParamsRegexProvider) }

var SplitParamsRegexProvider = validators.NewRegexpProvider(regexSplitParams, "splitParams", "")

func init() { validation.Register(Tiger128RegexProvider) }

var Tiger128RegexProvider = validators.NewRegexpProvider(regexTiger128, "tiger128", "")

func init() { validation.Register(Tiger160RegexProvider) }

var Tiger160RegexProvider = validators.NewRegexpProvider(regexTiger160, "tiger160", "")

func init() { validation.Register(Tiger192RegexProvider) }

var Tiger192RegexProvider = validators.NewRegexpProvider(regexTiger192, "tiger192", "")

func init() { validation.Register(ULIDRegexProvider) }

var ULIDRegexProvider = validators.NewRegexpProvider(regexULID, "ulid", "")

func init() { validation.Register(URLEncodedRegexProvider) }

var URLEncodedRegexProvider = validators.NewRegexpProvider(regexURLEncoded, "urlEncoded", "")

func init() { validation.Register(UuidRegexProvider) }

var UuidRegexProvider = validators.NewRegexpProvider(regexUuid, "uuid", "")

func init() { validation.Register(Uuid3RegexProvider) }

var Uuid3RegexProvider = validators.NewRegexpProvider(regexUuid3, "uuid3", "")

func init() { validation.Register(Uuid3RFC4122RegexProvider) }

var Uuid3RFC4122RegexProvider = validators.NewRegexpProvider(regexUuid3RFC4122, "uuid3Rfc4122", "")

func init() { validation.Register(Uuid4RegexProvider) }

var Uuid4RegexProvider = validators.NewRegexpProvider(regexUuid4, "uuid4", "")

func init() { validation.Register(Uuid4RFC4122RegexProvider) }

var Uuid4RFC4122RegexProvider = validators.NewRegexpProvider(regexUuid4RFC4122, "uuid4Rfc4122", "")

func init() { validation.Register(Uuid5RegexProvider) }

var Uuid5RegexProvider = validators.NewRegexpProvider(regexUuid5, "uuid5", "")

func init() { validation.Register(Uuid5RFC4122RegexProvider) }

var Uuid5RFC4122RegexProvider = validators.NewRegexpProvider(regexUuid5RFC4122, "uuid5Rfc4122", "")

func init() { validation.Register(UuidRFC4122RegexProvider) }

var UuidRFC4122RegexProvider = validators.NewRegexpProvider(regexUuidRFC4122, "uuidRfc4122", "")
