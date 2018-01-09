// +build !windows

package cli

func detectBinaryCandidates(bin string) ([]string, error) {
	bins := []string{"gpg2", "gpg1", "gpg"}
	if bin != "" {
		bins = append(bins, bin)
	}
	return bins, nil
}
