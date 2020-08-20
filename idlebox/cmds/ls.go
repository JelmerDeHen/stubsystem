package cmds
import (
	"os"
	"fmt"
	"syscall"
	"time"
)

func Ls(dirs []string) int {
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	for _, dir := range dirs {
//		files := make([]os.FileInfo, 0)
		dot, _ := os.Lstat(".")
		dotdot, _ := os.Lstat("..")
		files := []os.FileInfo{dot, dotdot}
		f, err := os.Open(dir)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer f.Close()
		fileInfo, err := f.Readdir(-1)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, file := range fileInfo {
			files = append(files, file)
		}

		//files, err := filepath.Glob(dir)
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
		//fmt.Println(files)
		fmt.Println("Mode\t\tNlink\tUid\tGid\tSize\tYYYY\tM\tD\tT\tName")
		for _, fi := range files {
//			fmt.Println(fi)
			sysinfo := fi.Sys().(*syscall.Stat_t)
			sec, nsec := sysinfo.Ctim.Unix()
			ts := time.Unix(sec, nsec)
			y,m,d := ts.Date()
			fmt.Printf("%s\t%d\t%d\t%d\t%d\t%d\t%s\t%02d\t%02d:%02d\t%s\n", fi.Mode().String(), sysinfo.Nlink, sysinfo.Uid, sysinfo.Gid, fi.Size(), y, m.String()[:3], d, ts.Hour(), ts.Minute(), fi.Name())
			//fmt.Printf("%v\n", sysinfo)
		}
	}
	return 0
}
