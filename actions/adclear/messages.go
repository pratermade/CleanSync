package adclear

type RemoveAdsMsg struct {
	started     bool
	tmpLocation string
	done        bool
	err         error
}

type CopyFileMsg struct {
	started     bool
	done        bool
	tmpLocation string
	err         error
}

type CleanTmpMsg struct {
}
