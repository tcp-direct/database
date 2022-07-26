package kv

import (
	"testing"

	c "git.tcp.direct/kayos/common/entropy"
)

func Test_Equal(t *testing.T) {
	t.Run("ShouldBeEqual", func(t *testing.T) {
		v := c.RandStr(55)
		kv1 := NewKeyValue(NewKey([]byte(v)), NewValue([]byte(v)))
		kv2 := NewKeyValue(NewKey([]byte(v)), NewValue([]byte(v)))
		if !kv1.Key.Equal(kv2.Key) {
			t.Fatalf("[FAIL] Keys not equal: %s, %s", kv1.Key.String(), kv2.Key.String())
		}
		if kv1.Key.String() != kv2.Key.String() {
			t.Fatalf(
				"[FAIL] Equal() passed but strings are not the same! kv1: %s != kv2: %s",
				kv1.Key.String(), kv2.Key.String())
		}
		if !kv1.Equal(kv2) {
			t.Fatal("[FAIL] KeyValue.Equal failed")
		}
		t.Logf("[+] KeyValues are equal: %s == %s", kv1.String(), kv2.String())

		if !kv1.Value.Equal(kv2.Value) {
			t.Fatalf("[FAIL] Values not equal: %s, %s", kv1.Value.String(), kv2.Value.String())
		}
		if kv1.Value.String() == kv2.Value.String() {
			t.Logf("[+] Values are equal: %s", kv1.Value.String())
		} else {
			t.Errorf(
				"[FAIL] Equal() passed but strings are not the same! kv1: %s != kv2: %s",
				kv1.Value.String(), kv2.Value.String(),
			)
		}
	})
	t.Run("ShouldNotBeEqual", func(t *testing.T) {
		v1 := c.RandStr(55)
		v2 := c.RandStr(55)
		kv1 := NewKeyValue(NewKey([]byte(v1)), NewValue([]byte(v1)))
		kv2 := NewKeyValue(NewKey([]byte(v2)), NewValue([]byte(v2)))
		if kv1.Key.Equal(kv2.Key) {
			t.Fatalf("[FAIL] Keys are equal: %s, %s", kv1.Key.String(), kv2.Key.String())
		}
		if kv1.Key.String() == kv2.Key.String() {
			t.Fatalf("[FAIL] Equal() did not pass but strings are the same! kv1: %s == kv2: %s",
				kv1.Key.String(), kv2.Key.String())
		}
		t.Logf("[+] Keys are not equal: %s", kv1.Key.String())
		if kv1.Value.Equal(kv2.Value) {
			t.Fatalf("[FAIL] Values are equal: %s, %s", kv1.Value.String(), kv2.Value.String())
		}
		if kv1.Value.String() == kv2.Value.String() {
			t.Fatalf("[FAIL] Equal() passed but strings are the same! kv1: %s != kv2: %s",
				kv1.Value.String(), kv2.Value.String())
		}
		t.Logf("[+] Values are not equal: %s", kv1.Value.String())
	})
}
