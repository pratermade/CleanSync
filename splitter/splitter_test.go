package splitter

// import (
// 	"testing"
// )

// func TestSplitFile(t *testing.T) {
// 	org := "X:\\shows\\Battlestar Galactica (2004)\\Season 4\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4"
// 	si := &SplitInfo{
// 		count:  0,
// 		offset: 0,
// 	}
// 	for {
// 		info, err := SplitFile(org, si)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		if info.eof {
// 			break
// 		}
// 		si.count = info.count
// 		si.offset = info.offset
// 		si.tempFolder = info.tempFolder
// 	}

// 	t.Log(si.tempFolder)

// }

// func TestRecombineFile(t *testing.T) {
// 	prefix := "C:\\Users\\pratermade\\AppData\\Local\\Temp\\s3sync936122437\\Battlestar Galactica (2003)  S04e19e20  Daybreak (1080P Bluray X265 Rzerox)-1.mp4"
// 	res, err := RecombineFile(prefix)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Fatal(res)

// }
