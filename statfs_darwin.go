//go:build darwin

package main

import "syscall"

func statfsFor(path string) (fsStats, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return fsStats{}, err
	}
	bs := uint64(st.Bsize)
	return fsStats{
		Total: st.Blocks * bs,
		Used:  (st.Blocks - st.Bfree) * bs,
		Free:  st.Bavail * bs,
	}, nil
}
