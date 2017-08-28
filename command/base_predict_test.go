package command

import (
	"reflect"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/posener/complete"
)

func TestPredictVaultPaths(t *testing.T) {
	t.Parallel()

	client, closer := testVaultServer(t)
	defer closer()

	data := map[string]interface{}{"a": "b"}
	if _, err := client.Logical().Write("secret/bar", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/foo", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/zip/zap", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/zip/zonk", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/zip/twoot", data); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		args         complete.Args
		includeFiles bool
		exp          []string
	}{
		{
			"has_args",
			complete.Args{
				All:  []string{"read", "secret/foo", "a=b"},
				Last: "a=b",
			},
			true,
			nil,
		},
		{
			"has_args_no_files",
			complete.Args{
				All:  []string{"read", "secret/foo", "a=b"},
				Last: "a=b",
			},
			false,
			nil,
		},
		{
			"part_mount",
			complete.Args{
				All:  []string{"read", "s"},
				Last: "s",
			},
			true,
			[]string{"secret/", "sys/"},
		},
		{
			"part_mount_no_files",
			complete.Args{
				All:  []string{"read", "s"},
				Last: "s",
			},
			false,
			[]string{"secret/", "sys/"},
		},
		{
			"only_mount",
			complete.Args{
				All:  []string{"read", "sec"},
				Last: "sec",
			},
			true,
			[]string{"secret/bar", "secret/foo", "secret/zip/"},
		},
		{
			"only_mount_no_files",
			complete.Args{
				All:  []string{"read", "sec"},
				Last: "sec",
			},
			false,
			[]string{"secret/zip/"},
		},
		{
			"full_mount",
			complete.Args{
				All:  []string{"read", "secret"},
				Last: "secret",
			},
			true,
			[]string{"secret/bar", "secret/foo", "secret/zip/"},
		},
		{
			"full_mount_no_files",
			complete.Args{
				All:  []string{"read", "secret"},
				Last: "secret",
			},
			false,
			[]string{"secret/zip/"},
		},
		{
			"full_mount_slash",
			complete.Args{
				All:  []string{"read", "secret/"},
				Last: "secret/",
			},
			true,
			[]string{"secret/bar", "secret/foo", "secret/zip/"},
		},
		{
			"full_mount_slash_no_files",
			complete.Args{
				All:  []string{"read", "secret/"},
				Last: "secret/",
			},
			false,
			[]string{"secret/zip/"},
		},
		{
			"path_partial",
			complete.Args{
				All:  []string{"read", "secret/z"},
				Last: "secret/z",
			},
			true,
			[]string{"secret/zip/twoot", "secret/zip/zap", "secret/zip/zonk"},
		},
		{
			"path_partial_no_files",
			complete.Args{
				All:  []string{"read", "secret/z"},
				Last: "secret/z",
			},
			false,
			[]string{"secret/zip/"},
		},
		{
			"subpath_partial_z",
			complete.Args{
				All:  []string{"read", "secret/zip/z"},
				Last: "secret/zip/z",
			},
			true,
			[]string{"secret/zip/zap", "secret/zip/zonk"},
		},
		{
			"subpath_partial_z_no_files",
			complete.Args{
				All:  []string{"read", "secret/zip/z"},
				Last: "secret/zip/z",
			},
			false,
			[]string{"secret/zip/z"},
		},
		{
			"subpath_partial_t",
			complete.Args{
				All:  []string{"read", "secret/zip/t"},
				Last: "secret/zip/t",
			},
			true,
			[]string{"secret/zip/twoot"},
		},
		{
			"subpath_partial_t_no_files",
			complete.Args{
				All:  []string{"read", "secret/zip/t"},
				Last: "secret/zip/t",
			},
			false,
			[]string{"secret/zip/t"},
		},
	}

	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				f := predictVaultPaths(client, tc.includeFiles)
				act := f(tc.args)
				if !reflect.DeepEqual(act, tc.exp) {
					t.Errorf("expected %q to be %q", act, tc.exp)
				}
			})
		}
	})
}

func TestPredictMounts(t *testing.T) {
	t.Parallel()

	client, closer := testVaultServer(t)
	defer closer()

	cases := []struct {
		name string
		path string
		exp  []string
	}{
		{
			"no_match",
			"not-a-real-mount-seriously",
			nil,
		},
		{
			"s",
			"s",
			[]string{"secret/", "sys/"},
		},
		{
			"se",
			"se",
			[]string{"secret/"},
		},
	}

	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				act := predictMounts(client, tc.path)
				if !reflect.DeepEqual(act, tc.exp) {
					t.Errorf("expected %q to be %q", act, tc.exp)
				}
			})
		}
	})
}

func TestPredictPaths(t *testing.T) {
	t.Parallel()

	client, closer := testVaultServer(t)
	defer closer()

	data := map[string]interface{}{"a": "b"}
	if _, err := client.Logical().Write("secret/bar", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/foo", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/zip/zap", data); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name         string
		path         string
		includeFiles bool
		exp          []string
	}{
		{
			"bad_path",
			"nope/not/a/real/path/ever",
			true,
			[]string{"nope/not/a/real/path/ever"},
		},
		{
			"good_path",
			"secret/",
			true,
			[]string{"secret/bar", "secret/foo", "secret/zip/"},
		},
		{
			"good_path_no_files",
			"secret/",
			false,
			[]string{"secret/zip/"},
		},
		{
			"partial_match",
			"secret/z",
			true,
			[]string{"secret/zip/"},
		},
		{
			"partial_match_no_files",
			"secret/z",
			false,
			[]string{"secret/zip/"},
		},
	}

	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				act := predictPaths(client, tc.path, tc.includeFiles)
				if !reflect.DeepEqual(act, tc.exp) {
					t.Errorf("expected %q to be %q", act, tc.exp)
				}
			})
		}
	})
}

func TestPredictListMounts(t *testing.T) {
	t.Parallel()

	client, closer := testVaultServer(t)
	defer closer()

	cases := []struct {
		name   string
		client *api.Client
		exp    []string
	}{
		{
			"not_connected_client",
			func() *api.Client {
				// Bad API client
				client, _ := api.NewClient(nil)
				return client
			}(),
			defaultPredictVaultMounts,
		},
		{
			"good_path",
			client,
			[]string{"cubbyhole/", "secret/", "sys/"},
		},
	}

	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				act := predictListMounts(tc.client)
				if !reflect.DeepEqual(act, tc.exp) {
					t.Errorf("expected %q to be %q", act, tc.exp)
				}
			})
		}
	})
}

func TestPredictListPaths(t *testing.T) {
	t.Parallel()

	client, closer := testVaultServer(t)
	defer closer()

	data := map[string]interface{}{"a": "b"}
	if _, err := client.Logical().Write("secret/bar", data); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Logical().Write("secret/foo", data); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		client *api.Client
		path   string
		exp    []string
	}{
		{
			"bad_path",
			client,
			"nope/not/a/real/path/ever",
			nil,
		},
		{
			"good_path",
			client,
			"secret/",
			[]string{"bar", "foo"},
		},
	}

	t.Run("group", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				act := predictListPaths(tc.client, tc.path)
				if !reflect.DeepEqual(act, tc.exp) {
					t.Errorf("expected %q to be %q", act, tc.exp)
				}
			})
		}
	})
}

func TestPredictHasPathArg(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
		exp  bool
	}{
		{
			"nil",
			nil,
			false,
		},
		{
			"empty",
			[]string{},
			false,
		},
		{
			"empty_string",
			[]string{""},
			false,
		},
		{
			"single",
			[]string{"foo"},
			false,
		},
		{
			"multiple",
			[]string{"foo", "bar", "baz"},
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if act := predictHasPathArg(tc.args); act != tc.exp {
				t.Errorf("expected %t to be %t", act, tc.exp)
			}
		})
	}
}