package main
import (
	"os/exec"
	"bytes"
	"io/ioutil"

	"text/tabwriter"
	"os"
	"fmt"
	"strings"
)
/*
kmod runs depmod: https://github.com/torvalds/linux/blob/3243a89dcbd8f5810b72ee0903d349bd000c4c9d/scripts/depmod.sh
This provides /lib/modules/${kver}/modules.dep{,.bin} based on build/System.map

kmod also supplies modinfo, this can read the binary format of modules.dep.bin and request path of module as such:
$ modinfo -k 5.6.12-arch1-1 -n iwlwifi
/lib/modules/5.6.12-arch1-1/kernel/drivers/net/wireless/intel/iwlwifi/iwlwifi.ko.xz

https://github.com/StarchLinux/kmod/blob/master/tools/modinfo.c#L177
-n param sets field="filename" calls kmod_module_get_path(mod)

line = kmod_search_moddep(mod->ctx, mod->name);
...
snprintf(fn, sizeof(fn), "%s/%s.bin", ctx->dirname,
					index_files[KMOD_INDEX_MODULES_DEP].fn);
idx = index_file_open(fn);
line = index_search(idx, name);
value = index_search__node(root, key, 0);


index_file_close(idx);
https://github.com/StarchLinux/kmod/blob/08515193d0ccf6c7b873e511dafe4b421b11376f/libkmod/libkmod-index.c#L442

*/
func getModulePath(kver, module string) (path string, err error) {
	if strings.HasPrefix(kver, "/") {
		// The full /lib/modules/... path was supplied
		kverSplit := strings.Split(kver, "/")
		if len(kverSplit) > 0 {
			kver = kverSplit[len(kverSplit)-1]
		}
	}
	koPath, err := exec.Command("modinfo", "-nk", kver, module).Output()
	if err != nil {
		return "", err
	}
	koPath = bytes.TrimSpace(koPath)
	return string(koPath), nil
}

func getLoadedModules() (modules []string, err error) {
	buf, err := ioutil.ReadFile("/proc/modules")
	if err != nil {
		return nil, err
	}
	for _, line := range(bytes.Split(buf, []byte{'\n'})) {
		spaceSplit := bytes.SplitN(line, []byte{' '}, 2)
		if len(spaceSplit) == 2 {
			modules = append(modules, string(spaceSplit[0]))
		}

	}
	return modules, nil
}

func LsmodEx() {
	mods := Lsmod("/sys/module")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	for _, mod := range mods {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", mod.Name, mod.Size, mod.Refcnt, strings.Join(mod.Usedby, ","))
		//GetModulePath(kver, mod.Name)
	}
	w.Flush()
}
