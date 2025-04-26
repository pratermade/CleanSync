package processVideo

type Status int

const (
	Starting    Status = iota // 0
	RemovingAds               // 1
	Uploading                 // 2
	Idle                      // 3
	Completed                 // 4
)

type ProcessVideoMessage struct {
	ndx         int
	status      Status
	tmpLocation string
	lastAction  string
	nextAction  string
	err         error
}

// type RemoveAdsMsg struct {
// 	started     bool
// 	tmpLocation string
// 	done        bool
// 	err         error
// }

// type CopyFileMsg struct {
// 	started     bool
// 	done        bool
// 	tmpLocation string
// 	err         error
// }

type CleanTmpMsg struct {
}
