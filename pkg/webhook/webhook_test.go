package webhook

// FIXME: need test

// func TestCreatePatchForPodh(t *testing.T) {

// 	testCases := []struct {
// 		description string
// 		file        string
// 		output      string
// 		expected    bool
// 	}{
// 		{
// 			description: "toleration and taint have the same key and effect, and operator is Exists, and taint has no value, expect tolerated",
// 			file:        "testdata/pod.json",
// 			output:      `[{"op":"replace","path":"/metadata/annotations/mac-compute-node-selector","value":"injected"},{"op":"replace","path":"/spec/spec/nodeSelector","value":{"nodeSelectorTerms":[{"matchFields":[{"key":"kubernetes.io/arch","operator":"In","values":["amd64"]},{"key":"node.openshift.io/os_id","operator":"In","values":["rhcos"]}]}]}}]`,
// 			expected:    true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		data, err := os.ReadFile(tc.file)
// 		if err != nil {
// 			t.Errorf("error serializing body %v", err)
// 		}

// 		var pod v12.Pod
// 		if err := json.Unmarshal(data, &pod); err != nil {
// 			t.Errorf("Unable to serialize")
// 		}

// 		out, err := CreatePatchForPod(&pod)

// 		fmt.Println(string(out))
// 		val := string(out) == tc.output
// 		if val != tc.expected {
// 			t.Errorf("[%s] expect %v, got %v: ", tc.description, tc.expected, val)
// 		}
// 	}
// }
