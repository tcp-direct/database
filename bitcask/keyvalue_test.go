package bitcask

import (
	"testing"

	"git.tcp.direct/kayos/common/entropy"
)

func Test_Equal(t *testing.T) {
	t.Run("ShouldBeEqual", func(t *testing.T) {
		v := entropy.RandStr(55)
		kv1 := KeyValue{Key{b: []byte(v)}, Value{b: []byte(v)}}
		kv2 := KeyValue{Key{b: []byte(v)}, Value{b: []byte(v)}}
		if !kv1.Key.Equal(kv2.Key) {
			t.Errorf("[FAIL] Keys not equal: %s, %s", kv1.Key.String(), kv2.Key.String())
		} else {
			if kv1.Key.String() == kv2.Key.String() {
				t.Logf("[+] Keys are equal: %s", kv1.Key.String())
			} else {
				t.Errorf(
					"[FAIL] Equal() passed but strings are not the same! kv1: %s != kv2: %s",
					kv1.Key.String(), kv2.Key.String(),
				)
			}
		}

		if !kv1.Value.Equal(kv2.Value) {
			t.Errorf("[FAIL] Values not equal: %s, %s", kv1.Value.String(), kv2.Value.String())
		} else {
			if kv1.Value.String() == kv2.Value.String() {
				t.Logf("[+] Values are equal: %s", kv1.Value.String())
			} else {
				t.Errorf(
					"[FAIL] Equal() passed but strings are not the same! kv1: %s != kv2: %s",
					kv1.Value.String(), kv2.Value.String(),
				)
			}

		}
	})

	t.Run("ShouldNotBeEqual", func(t *testing.T) {
		v1 := entropy.RandStr(55)
		v2 := entropy.RandStr(55)
		kv1 := KeyValue{Key{b: []byte(v1)}, Value{b: []byte(v1)}}
		kv2 := KeyValue{Key{b: []byte(v2)}, Value{b: []byte(v2)}}
		if kv1.Key.Equal(kv2.Key) {
			t.Errorf("[FAIL] Keys are equal: %s, %s", kv1.Key.String(), kv2.Key.String())
		} else {
			if kv1.Key.String() != kv2.Key.String() {
				t.Logf("[+] Keys are not equal: %s", kv1.Key.String())
			} else {
				t.Errorf(
					"[FAIL] Equal() passed but strings are the same! kv1: %s != kv2: %s",
					kv1.Key.String(), kv2.Key.String(),
				)
			}
		}

		if kv1.Value.Equal(kv2.Value) {
			t.Errorf("[FAIL] Values are equal: %s, %s", kv1.Value.String(), kv2.Value.String())
		} else {
			if kv1.Value.String() != kv2.Value.String() {
				t.Logf("[+] Values are not equal: %s", kv1.Value.String())
			} else {
				t.Errorf(
					"[FAIL] Equal() passed but strings are the same! kv1: %s != kv2: %s",
					kv1.Value.String(), kv2.Value.String(),
				)
			}

		}
	})

}
