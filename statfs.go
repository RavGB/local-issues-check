package main

// fsStats holds the filesystem figures the disk check needs, in
// bytes. Free counts the space available to an ordinary user (the
// statfs Bavail figure), and Used is Total minus the physically free
// blocks, so Total, Free and Used always tell a consistent story.
type fsStats struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// statfsFor is implemented per platform in statfs_linux.go,
// statfs_darwin.go and statfs_other.go.
