package control

import (
	"fmt"
	"time"

	capture "github.com/ygelfand/libcountertop/pkg/audio/capture"

	"github.com/spf13/cobra"

	"github.com/ygelfand/echolocal/internal/hardware/mic"
)

const (
	shortest = time.Second
	longest  = 60 * time.Second
)

func hearing() []*cobra.Command {
	record := &cobra.Command{
		Use:   "record <seconds> [path]",
		Short: "Record what wake detection hears, and say how loud it was",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := seconds(args[0])
			if err != nil {
				return err
			}

			frames, stop := mic.Get().Listen("ctl audio record")
			defer stop()

			var got []int16
			err = capture.Listen(
				cmd.Context(),
				d,
				frames,
				func(f []int16) { got = append(got, f...) },
			)
			if err != nil {
				return err
			}
			if len(got) == 0 {
				return fmt.Errorf("record: nothing captured")
			}

			say := fmt.Sprintf("seconds %.2f %s", float64(len(got))/mic.Rate, levels(got))
			if len(args) > 1 {
				if err := write(args[1], 1, got); err != nil {
					return err
				}
				say = args[1] + " " + say
			}
			fmt.Fprintln(cmd.OutOrStdout(), say)
			return nil
		},
	}

	echo := &cobra.Command{
		Use:   "echo <seconds> <path>",
		Short: "Record the center microphone, the loopback, and what listeners got, as three channels",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := seconds(args[0])
			if err != nil {
				return err
			}

			frames, stop := mic.Get().ListenEcho()
			defer stop()

			var m, r, o []int16
			var sounding int
			err = capture.Listen(cmd.Context(), d, frames, func(e mic.Echo) {
				n := min(len(e.Mic), len(e.Ref), len(e.Out))
				m, r, o = append(m, e.Mic[:n]...), append(r, e.Ref[:n]...), append(o, e.Out[:n]...)
				if e.Sounding {
					sounding += n
				}
			})
			if err != nil {
				return err
			}
			if len(m) == 0 {
				return fmt.Errorf("record echo: nothing captured")
			}

			wav := make([]int16, 0, 3*len(m))
			for i := range m {
				wav = append(wav, m[i], r[i], o[i])
			}
			if err := write(args[1], 3, wav); err != nil {
				return err
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s seconds %.2f sounding %.2f\n  mic %s\n  ref %s\n  out %s\n",
				args[1],
				float64(len(m))/mic.Rate,
				float64(sounding)/mic.Rate,
				levels(m),
				levels(r),
				levels(o),
			)
			return nil
		},
	}

	record.AddCommand(echo)
	return []*cobra.Command{record}
}

func seconds(arg string) (time.Duration, error) { return capture.Duration(arg, shortest, longest) }
func levels(samples []int16) string             { return capture.Measure(samples).String() }

func write(path string, channels int, samples []int16) error {
	return capture.Write(path, samples, mic.Rate, channels)
}
