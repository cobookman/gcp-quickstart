package renders

import (
  "os"
  "errors"
  "path/filepath"
  "github.com/cobookman/gcp-quickstart/layout"
  "strings"
  "os/exec"
)

// Builds a claat source
func RenderClaat(lesson *layout.Lesson, ga string, buildFolder string) error {
  if len(lesson.SourceClaat) == 0 {
    return errors.New("No claat source given")
  }
  // Where we are building the claat
  buildPath := strings.Replace(lesson.Href, "index.html", "", 1)

  // Create a temporary scratch location
	scratchFolder := filepath.Join("scratch/",  buildPath)

	os.RemoveAll(scratchFolder)
	os.Mkdir(scratchFolder, os.ModePerm)

	defer func() {
		os.RemoveAll(scratchFolder)
	}()

	// Render Claat
	claatCmd := exec.Command("./claat",
		"export",
    "-prefix", "/",
		"-f", "html",
		"-ga", ga,
		"-o", scratchFolder,
		lesson.SourceClaat)

	claatCmd.Stdout = os.Stdout
	claatCmd.Stderr = os.Stderr
	if err := claatCmd.Start(); err != nil {
		return err
	}
	if err := claatCmd.Wait(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(buildFolder, buildPath), os.ModePerm); err != nil {
		return err
	}
	claatName, _ := filepath.Glob(scratchFolder + "/*")
	copyCmd := exec.Command("cp", "-R", string(claatName[0]) + "/", filepath.Join(buildFolder, buildPath))
	copyCmd.Stdout = os.Stdout
	copyCmd.Stderr = os.Stderr
	if err := copyCmd.Start(); err != nil {
		return err
	}
	if err := copyCmd.Wait(); err != nil {
		return err
	}

	return nil
}
