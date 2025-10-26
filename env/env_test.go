package env

import (
	"testing"
)

func TestString(t *testing.T) {
	t.Run("string with no default", func(t *testing.T) {
		envs := map[string]string{"envvar": "myval"}
		setupEnvVars(t, envs)
		if got, want := String("envvar"), "myval"; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("string with default - env var found", func(t *testing.T) {
		envs := map[string]string{"envvar": "myval"}
		setupEnvVars(t, envs)
		if got, want := StringWithDefault("envvar", "myval2"), "myval"; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("string with default - env var not found", func(t *testing.T) {
		if got, want := StringWithDefault("envvar", "myval2"), "myval2"; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
}

func TestInt(t *testing.T) {
	t.Run("int with no default", func(t *testing.T) {
		envs := map[string]string{"envvar": "1212"}
		setupEnvVars(t, envs)
		if got, want := Int("envvar"), 1212; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("int with default - env var found", func(t *testing.T) {
		envs := map[string]string{"envvar": "1212"}
		setupEnvVars(t, envs)
		if got, want := IntWithDefault("envvar", 1111), 1212; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("int with default - env var not int", func(t *testing.T) {
		envs := map[string]string{"envvar": "121a"}
		setupEnvVars(t, envs)
		if got, want := IntWithDefault("envvar", 1111), 1111; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("int with default - env var not found", func(t *testing.T) {
		if got, want := IntWithDefault("envvar", 1111), 1111; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
}

func TestBool(t *testing.T) {
	t.Run("bool with no default", func(t *testing.T) {
		envs := map[string]string{"envvar": "true"}
		setupEnvVars(t, envs)
		if got, want := Bool("envvar"), true; got != want {
			t.Errorf("got a different value than the wanted one. expected: %t; got: %t", want, got)
		}
	})
	t.Run("bool with default - env var found", func(t *testing.T) {
		envs := map[string]string{"envvar": "false"}
		setupEnvVars(t, envs)
		if got, want := BoolWithDefault("envvar", true), false; got != want {
			t.Errorf("got a different value than the wanted one. expected: %t; got: %t", want, got)
		}
	})
	t.Run("bool with default - env var not bool", func(t *testing.T) {
		envs := map[string]string{"envvar": "test"}
		setupEnvVars(t, envs)
		if got, want := BoolWithDefault("envvar", true), true; got != want {
			t.Errorf("got a different value than the wanted one. expected: %t; got: %t", want, got)
		}
	})
	t.Run("bool with default - env var not found", func(t *testing.T) {
		if got, want := BoolWithDefault("envvar", true), true; got != want {
			t.Errorf("got a different value than the wanted one. expected: %t; got: %t", want, got)
		}
	})
}

func TestExpand(t *testing.T) {
	t.Run("expand with ref", func(t *testing.T) {
		envs := map[string]string{
			"env1": "env1 val",
			"env2": "env 2 contains also ${env1}",
		}
		setupEnvVars(t, envs)
		if got, want := Expand(String("env2")), "env 2 contains also env1 val"; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
	t.Run("expand with ref on non existing", func(t *testing.T) {
		envs := map[string]string{
			"env2": "env 2 contains also ${env1}",
		}
		setupEnvVars(t, envs)
		if got, want := Expand(String("env2")), "env 2 contains also "; got != want {
			t.Errorf("got a different value than the wanted one. expected: %q; got: %q", want, got)
		}
	})
}

func setupEnvVars(t *testing.T, in map[string]string) {
	for k, v := range in {
		t.Setenv(k, v)
	}
}
