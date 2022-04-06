package tar

type Packager struct{}

func NewPackager() Packager {
	return Packager{}
}

func (Packager) Package(sourceFolder, dstFile string) error {
	return TarFolder(sourceFolder, dstFile)
}

func (Packager) UnPackage(orgFile, dstFolder string) error {
	return UntarFile(orgFile, dstFolder)
}
