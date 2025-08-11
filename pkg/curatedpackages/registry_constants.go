package curatedpackages

const (
	devNonRegionalPublicRegistryAlias = "l0g8r8j6"
	devRegionalPublicRegistryAlias    = "x3k6m8v0"
	stagingPublicRegistryAlias        = "w9m0f3l5"
	prodPublicRegistryAlias           = "eks-anywhere"
	devNonRegionalPublicRegistryURI   = "public.ecr.aws/" + devNonRegionalPublicRegistryAlias
	devRegionalPublicRegistryURI      = "public.ecr.aws/" + devRegionalPublicRegistryAlias
	stagingPublicRegistryURI          = "public.ecr.aws/" + stagingPublicRegistryAlias
	prodPublicRegistryURI             = "public.ecr.aws/" + prodPublicRegistryAlias
	prodNonRegionalPrivateRegistryURI = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
	devRegionalPrivateRegistryURI     = "067575901363.dkr.ecr.us-west-2.amazonaws.com"
	stagingRegionalPrivateRegistryURI = "724423470321.dkr.ecr.us-west-2.amazonaws.com"
)

var prodRegionalPrivateRegistryURIByRegion = map[string]string{
	"af-south-1":     "783635962247.dkr.ecr.af-south-1.amazonaws.com",
	"ap-east-1":      "804323328300.dkr.ecr.ap-east-1.amazonaws.com",
	"ap-northeast-1": "143143237519.dkr.ecr.ap-northeast-1.amazonaws.com",
	"ap-northeast-2": "447311122189.dkr.ecr.ap-northeast-2.amazonaws.com",
	"ap-northeast-3": "376465423944.dkr.ecr.ap-northeast-3.amazonaws.com",
	"ap-south-1":     "357015164304.dkr.ecr.ap-south-1.amazonaws.com",
	"ap-south-2":     "388483641499.dkr.ecr.ap-south-2.amazonaws.com",
	"ap-southeast-1": "654894141437.dkr.ecr.ap-southeast-1.amazonaws.com",
	"ap-southeast-2": "299286866837.dkr.ecr.ap-southeast-2.amazonaws.com",
	"ap-southeast-3": "703305448174.dkr.ecr.ap-southeast-3.amazonaws.com",
	"ap-southeast-4": "106475008004.dkr.ecr.ap-southeast-4.amazonaws.com",
	"ap-southeast-5": "396913739932.dkr.ecr.ap-southeast-5.amazonaws.com",
	"ap-southeast-7": "141666480881.dkr.ecr.ap-southeast-7.amazonaws.com",
	"ca-central-1":   "064352486547.dkr.ecr.ca-central-1.amazonaws.com",
	"ca-west-1":      "571600859530.dkr.ecr.ca-west-1.amazonaws.com",
	"eu-central-1":   "364992945014.dkr.ecr.eu-central-1.amazonaws.com",
	"eu-central-2":   "551422459769.dkr.ecr.eu-central-2.amazonaws.com",
	"eu-north-1":     "826441621985.dkr.ecr.eu-north-1.amazonaws.com",
	"eu-south-1":     "787863792200.dkr.ecr.eu-south-1.amazonaws.com",
	"eu-south-2":     "127214161500.dkr.ecr.eu-south-2.amazonaws.com",
	"eu-west-1":      "090204409458.dkr.ecr.eu-west-1.amazonaws.com",
	"eu-west-2":      "371148654473.dkr.ecr.eu-west-2.amazonaws.com",
	"eu-west-3":      "282646289008.dkr.ecr.eu-west-3.amazonaws.com",
	"il-central-1":   "131750224677.dkr.ecr.il-central-1.amazonaws.com",
	"me-central-1":   "454241080883.dkr.ecr.me-central-1.amazonaws.com",
	"me-south-1":     "158698011868.dkr.ecr.me-south-1.amazonaws.com",
	"mx-central-1":   "295916874491.dkr.ecr.mx-central-1.amazonaws.com",
	"sa-east-1":      "517745584577.dkr.ecr.sa-east-1.amazonaws.com",
	"us-east-1":      "331113665574.dkr.ecr.us-east-1.amazonaws.com",
	"us-east-2":      "297090588151.dkr.ecr.us-east-2.amazonaws.com",
	"us-west-1":      "440460740297.dkr.ecr.us-west-1.amazonaws.com",
	"us-west-2":      "346438352937.dkr.ecr.us-west-2.amazonaws.com",
}
