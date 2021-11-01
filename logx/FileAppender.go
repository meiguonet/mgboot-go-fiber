package logx

import (
	"fmt"
	"github.com/meiguonet/mgboot-go-common/util/castx"
	"github.com/meiguonet/mgboot-go-common/util/fsx"
	"github.com/meiguonet/mgboot-go-common/util/stringx"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type fileAppender struct {
	channel   string
	filepath  string
	maxSize   int64
	maxBackup int
}

func newFileAppender(settings map[string]interface{}) *fileAppender {
	var channel string

	if s1, ok := settings["channel"].(string); ok {
		channel = s1
	}

	var fpath string

	if s1, ok := settings["filepath"].(string); ok {
		fpath = s1
	}

	maxSize := int64(20 * 1024 * 1024)

	if n1, ok := settings["maxSize"].(int64); ok && n1 > 0 {
		maxSize = n1
	} else if s1, ok := settings["maxSize"].(string); ok && s1 != "" {
		n1 := castx.ToDataSize(s1)

		if n1 > 0 {
			maxSize = n1
		}
	}

	maxBackup := 7

	if n1, ok := settings["maxBackup"].(int); ok && n1 > 0 {
		maxBackup = n1
	}

	return &fileAppender{
		channel:   channel,
		filepath:  fpath,
		maxSize:   maxSize,
		maxBackup: maxBackup,
	}
}

func (a *fileAppender) GetAppenderName() string {
	return "FileAppender"
}

func (a *fileAppender) Write(buf []byte) (int, error) {
	var fpath string

	if a.filepath != "" {
		fpath = fsx.GetRealpath(a.filepath)
		fpath = a.filepath
		fpath = strings.ReplaceAll(fpath, "\\", "/")

		if stat, err := os.Stat(a.filepath); err != nil || stat.IsDir() {
			fpath = ""
		}
	}

	if fpath == "" {
		if a.channel == "" {
			return len(buf), nil
		}

		dir := logDir

		if dir == "" {
			dir = fsx.GetRealpath("datadir:logs")
		}

		dir = strings.ReplaceAll(dir, "\\", "/")
		dir = strings.TrimRight(dir, "/")

		if dir == "" {
			return len(buf), nil
		}

		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			os.MkdirAll(dir, 0755)
		}

		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			return len(buf), nil
		}

		fpath = fmt.Sprintf("%s/%s.log", dir, a.channel)
	}

	if fpath == "" {
		return len(buf), nil
	}

	a.rollingFile(fpath)
	parts := strings.Split(string(buf), fieldSep)
	var ts, channel, level, msg string

	for _, p := range parts {
		if strings.HasPrefix(p, "ts:") {
			ts = strings.TrimPrefix(p, "ts:")
			continue
		}

		if strings.HasPrefix(p, "channel:") {
			channel = strings.TrimPrefix(p, "channel:")
			continue
		}

		if strings.HasPrefix(p, "level:") {
			level = strings.TrimPrefix(p, "level:")
			continue
		}

		if strings.HasPrefix(p, "msg:") {
			msg = strings.TrimPrefix(p, "msg:")
		}
	}

	_, msg = a.handleMsgTags(msg)
	sb := strings.Builder{}
	sb.WriteString("[")
	sb.WriteString(ts)
	sb.WriteString("]")

	if channel != "" {
		sb.WriteString("[")
		sb.WriteString(channel)
		sb.WriteString("]")
	}

	sb.WriteString("[")
	sb.WriteString(level)
	sb.WriteString("] ")
	sb.WriteString(msg)

	if fsx.IsWin() {
		sb.WriteString("\r\n")
	} else {
		sb.WriteString("\n")
	}

	ioutil.WriteFile(fpath, []byte(sb.String()), 0755)
	return len(buf), nil
}

func (a *fileAppender) rollingFile(fpath string) {
	stat, err := os.Stat(fpath)

	if err != nil || stat.IsDir() {
		return
	}

	if stat.Size() <= a.maxSize {
		return
	}

	buf, _ := ioutil.ReadFile(fpath)
	dir := filepath.Dir(fpath)
	dir = strings.ReplaceAll(dir, "\\", "/")
	dir = strings.TrimRight(dir, "/")
	fname := filepath.Base(fpath)
	fname = strings.TrimSuffix(fname, ".log")
	re1 := regexp.MustCompile(fname + `\.([0-9]+)\.log$`)
	backups := make([]map[string]interface{}, 0)

	filepath.Walk(dir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		matches := re1.FindStringSubmatch(backupPath)

		if len(matches) < 2 {
			return nil
		}

		backupPath = strings.ReplaceAll(backupPath, "\\", "/")
		backupPath = stringx.EnsureLeft(backupPath, dir + "/")
		contents, _ := ioutil.ReadFile(backupPath)
		os.Remove(backupPath)

		if len(contents) > 0 {
			backups = append(backups, map[string]interface{}{
				"idx":      castx.ToInt(matches[1]),
				"contents": contents,
			})
		}

		return nil
	})

	if len(backups) > 1 {
		sort.SliceStable(backups, func(i, j int) bool {
			return castx.ToInt(backups[i]["idx"]) < castx.ToInt(backups[j]["idx"])
		})
	}

	ioutil.WriteFile(fpath, []byte{}, 0755)
	ioutil.WriteFile(fmt.Sprintf("%s/%s.1.log", dir, fname), buf, 0755)

	if a.maxBackup > 1 {
		for i := 2; i <= a.maxBackup; i++ {
			if len(backups) < 1 {
				break
			}

			entry := backups[0]
			backups = backups[1:]
			ioutil.WriteFile(fmt.Sprintf("%s/%s.%d.log", dir, fname, i), entry["contents"].([]byte), 0755)
		}
	}
}

func (a *fileAppender) handleMsgTags(msg string) ([][]string, string) {
	re1 := regexp.MustCompile(`^tags[\x20\t]*=[\x20\t]*\[([^]]+)]`)
	tags := make([][]string, 0)
	groups := re1.FindStringSubmatch(msg)

	if len(groups) < 2 {
		return tags, strings.TrimSpace(msg)
	}

	msg = strings.TrimPrefix(msg, groups[0])
	msg = strings.TrimSpace(msg)
	re2 := regexp.MustCompile(`[\x20\t]+`)
	parts := re2.Split(strings.TrimSpace(groups[1]), -1)
	re3 := regexp.MustCompile(`[\x20\t]*:[\x20\t]*`)

	for _, p := range parts {
		a1 := re3.Split(p, -1)

		if len(a1) != 2 {
			continue
		}

		p1 := strings.TrimSpace(a1[0])
		p2 := strings.TrimSpace(a1[1])

		if p1 == "" || p2 == "" {
			continue
		}

		tags = append(tags, []string{p1, p2})
	}

	return tags, msg
}
