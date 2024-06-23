package test_explorer

import "strings"

type resolver interface {
	resolveTestingItem(pkgPath string, name string) (*TestingItem, error)
}

func buildEvent(testEvent *TestEvent, pathPrefix []string, modPath string, pm *pathMapping, r resolver) ([]*TestingItemEvent, error) {
	var item *TestingItem
	var status RunStatus

	switch testEvent.Action {
	case TestEventAction_Run:
		status = RunStatus_Running
	case TestEventAction_Pass:
		status = RunStatus_Success
	case TestEventAction_Fail:
		status = RunStatus_Fail
	case TestEventAction_Skip:
		status = RunStatus_Skip
	}

	var events []*TestingItemEvent

	var basePath []string
	var path []string
	if testEvent.Package != "" {
		if testEvent.Test != "" {
			var testSuffix string
			baseTest := testEvent.Test
			idx := strings.Index(testEvent.Test, "/")
			if idx >= 0 {
				baseTest = testEvent.Test[:idx]
				testSuffix = testEvent.Test[idx+1:]
			}

			// dynamically generated sub case
			item, _ = r.resolveTestingItem(testEvent.Package, baseTest)
			if item != nil {
				path = getCaseItemPath(pathPrefix, item.RelPath, baseTest, "")
				if testSuffix != "" {
					suffix := strings.Split(testSuffix, "/")
					basePath = path
					path = append(path, suffix...)

					// check if path is added to tree, if not, add a sync event to tell
					// the FE to add new sub tree
					ok, _ := pm.Get(path)
					if !ok {
						// item = root item
						// make a copy
						item = item.CloneSelf()
						item.Name = suffix[len(suffix)-1]
						item.BaseCaseName = baseTest
						item.NameUnderPkg = testEvent.Test
						item.Key = item.Name
						item.State = &TestingItemState{Status: status}

						// event type = MergeTree
						// merge operation:
						//     missing: add
						//     existing: merge with prev
						events = append(events, &TestingItemEvent{
							Event: Event_MergeTree,
							Item:  makeMergeTree(suffix[:len(suffix)-1], baseTest, item),
							Path:  basePath,
						})

						// set initial status
						for i := 0; i < len(suffix); i++ {
							events = append(events, &TestingItemEvent{
								Event:  Event_ItemStatus,
								Path:   append(basePath, suffix[:i+1]...),
								Status: RunStatus_Running,
							})
						}
					}
				}
			}
		} else {
			subPath := getPkgSubPath(modPath, testEvent.Package)
			path = appendCopy(pathPrefix)
			if subPath != "" {
				path = append(path, strings.Split(subPath, "/")...)
			}
		}
	}

	events = append(events, &TestingItemEvent{
		Event:  Event_ItemStatus,
		Item:   item,
		Path:   path,
		Status: status,
		Msg:    testEvent.Output,
	})
	if len(basePath) > 0 {
		// append output to base test
		events = append(events, &TestingItemEvent{
			Event: Event_ItemStatus,
			Path:  basePath,
			Msg:   testEvent.Output,
		})
	}
	return events, nil
}
