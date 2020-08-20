package main
import (
	"os"
	"errors"
	"fmt"
	"bytes"
)
// The procfs /proc/version contains compile-time release information

/*
 * Parses /proc/version format to []string{}
 * procVersion("/proc/version")
 */
func UnameR(path string) (string, error) {
	version, err := ProcVersionRead("/proc/version")
	if err != nil {
		return "", err
	}
	uname, err := ProcVersionProcess(version)
	if err != nil {
		return "", err
	}
	//PpUname(uname)
	return string(uname.Utsname__release), nil
}
// Separate the read & process functionality. This way its usable in a busy cmp loop
func ProcVersionRead(path string) ([]byte, error) {
	// Read the file
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	// Read /proc/version to fixed size buf
	buf := make([]byte, 1337)
	n, err := fh.Read(buf)
	if err != nil {
		return nil, err
	}
	if n == 1337 || n == 0 {
		return nil, errors.New("Error reading /proc/version")
	}
	return buf, nil
}
/*
https://github.com/torvalds/linux/blob/master/fs/proc/version.c
static int version_proc_show(struct seq_file *m, void *v)
{
	seq_printf(m, linux_proc_banner,
		utsname()->sysname,
		utsname()->release,
		utsname()->version);
	return 0;
}
https://github.com/torvalds/linux/blob/master/init/version.c
const char linux_proc_banner[] =
	"%s version %s"
	" (" LINUX_COMPILE_BY "@" LINUX_COMPILE_HOST ")"
	" (" LINUX_COMPILER ") %s\n";

https://github.com/torvalds/linux/blob/master/scripts/mkcompile_h
Generates compile.h with local variables
LINUX_COMPILE_BY=$(whoami | sed 's/\\/\\\\/')
LINUX_COMPILE_HOST=`hostname`

CC_VERSION=$($CC -v 2>&1 | grep ' version ' | sed 's/[[:space:]]*$//')
LD_VERSION=$($LD -v | head -n1 | sed 's/(compatible with [^)]*)//' \
		      | sed 's/[[:space:]]*$//')
printf '#define LINUX_COMPILER "%s"\n' "$CC_VERSION, $LD_VERSION"

*/
type Uname struct {
	Utsname__sysname []byte
	Utsname__release []byte
	LINUX_COMPILE_BY []byte
	LINUX_COMPILE_HOST []byte
	LINUX_COMPILER []byte
	Utsname__version []byte
}
func ProcVersionProcess(version []byte) (*Uname, error) {
	var buf []byte
	type UnameStage int
	const (
		Utsname__sysname UnameStage = iota
		Utsname__release
		LINUX_COMPILE_BY
		LINUX_COMPILE_HOST
		LINUX_COMPILER
		Utsname__version
	)
	var unameStage UnameStage = Utsname__sysname
	uname := &Uname{}
	// Loop through chars in /proc/version moving through stages as parts of the format string are recognized
	for _, b := range version {
		if b == 0 {
			break
		}
		buf = append(buf, b)
		switch unameStage {
		case Utsname__sysname:
			// Look for the magic "% version %"
			if bytes.HasSuffix(buf, []byte(" version ")) {
				uname.Utsname__sysname = bytes.TrimSuffix(buf, []byte(" version "))
				unameStage=Utsname__release
				buf = []byte{}
			}
		case Utsname__release:
			if bytes.HasSuffix(buf, []byte(" (")) {
				uname.Utsname__release = bytes.TrimSuffix(buf, []byte(" ("))
				unameStage=LINUX_COMPILE_BY
				buf = []byte{}
			}
		case LINUX_COMPILE_BY:
			if b == '@' {
				uname.LINUX_COMPILE_BY = bytes.TrimSuffix(buf, []byte{'@'})
				unameStage=LINUX_COMPILE_HOST
				buf = []byte{}
			}
		case LINUX_COMPILE_HOST:
			if bytes.HasSuffix(buf, []byte(") (")) {
				uname.LINUX_COMPILE_HOST = bytes.TrimSuffix(buf, []byte(") ("))
				unameStage=LINUX_COMPILER
				buf = []byte{}
			}
		case LINUX_COMPILER:
			if bytes.HasSuffix(buf, []byte(") ")) {
				uname.LINUX_COMPILER = bytes.TrimSuffix(buf, []byte(") "))
				unameStage=Utsname__version
				buf=[]byte{}
			}
		case Utsname__version:
			if b == '\n' {
				uname.Utsname__version=bytes.TrimSuffix(buf, []byte("\n"))
				break
			}
		default:
			fmt.Printf("Unknown stage %+v %s\n", unameStage, buf)
		}
	}
	
	//fmt.Println(string(buf), "yay")
	return uname, nil
}


/*
	Utsname__sysname []byte
	Utsname__release []byte
	LINUX_COMPILE_BY []byte
	LINUX_COMPILE_HOST []byte
	LINUX_COMPILER []byte
	Utsname__version []byte
*/
func PpUname (uname *Uname) {
	fmt.Printf("utsname()->sysname=%s\n", string(uname.Utsname__sysname))
	fmt.Printf("utsname()->release=%s\n", string(uname.Utsname__release))
	fmt.Printf("utsname()->version=%s\n", string(uname.Utsname__version))


	fmt.Printf("LINUX_COMPILE_BY=%s\n", string(uname.LINUX_COMPILE_BY))
	fmt.Printf("LINUX_COMPILE_HOST=%s\n", string(uname.LINUX_COMPILE_HOST))
	fmt.Printf("LINUX_COMPILER=%s\n", string(uname.LINUX_COMPILER))
}
