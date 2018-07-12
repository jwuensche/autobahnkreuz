package filter

import (
	"strings"

	"github.com/gammazero/nexus/router"
	"github.com/gammazero/nexus/wamp"
)

type simplePublishFilter struct {
	blIDs        []wamp.ID
	wlIDs        []wamp.ID
	blMap        map[string][]string
	wlMap        map[string][]string
	lockRequired bool
}

// NewSimplePublishFilter gets any blacklists and whitelists included in a PUBLISH
// message.  If there are no filters defined by the PUBLISH message, then nil
// is returned.
func NewSimplePublishFilter(opts wamp.Dict) router.PublishFilter {
	const (
		blacklistPrefix = "exclude_"
		whitelistPrefix = "eligible_"
	)

	if len(opts) == 0 {
		return nil
	}

	var blIDs []wamp.ID
	if blacklist, ok := wamp.AsList(opts[wamp.BlacklistKey]); ok {
		for i := range blacklist {
			if blVal, ok := wamp.AsID(blacklist[i]); ok {
				blIDs = append(blIDs, blVal)
			}
		}
	} else if blacklist, ok := wamp.AsID(opts[wamp.BlacklistKey]); ok {
		blIDs = append(blIDs, blacklist)
	}

	var wlIDs []wamp.ID
	if whitelist, ok := wamp.AsList(opts[wamp.WhitelistKey]); ok {
		for i := range whitelist {
			if wlID, ok := wamp.AsID(whitelist[i]); ok {
				wlIDs = append(wlIDs, wlID)
			}
		}
	} else if whitelist, ok := wamp.AsID(opts[wamp.WhitelistKey]); ok {
		wlIDs = append(wlIDs, whitelist)
	}

	getAttrMap := func(prefix string) map[string][]string {
		var attrMap map[string][]string
		for k, values := range opts {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			if vals, ok := wamp.AsList(values); ok {
				vallist := make([]string, 0, len(vals))
				for i := range vals {
					if val, ok := wamp.AsString(vals[i]); ok && val != "" {
						vallist = append(vallist, val)
					}
				}
				if len(vallist) != 0 {
					attrName := k[len(prefix):]
					if attrMap == nil {
						attrMap = map[string][]string{}
					}
					attrMap[attrName] = vallist
				}
			} else if val, ok := wamp.AsString(values); ok {
				attrName := k[len(prefix):]
				if attrMap == nil {
					attrMap = map[string][]string{}
				}
				attrMap[attrName] = []string{val}
			}
		}
		return attrMap
	}

	blMap := getAttrMap(blacklistPrefix)
	wlMap := getAttrMap(whitelistPrefix)

	if blIDs == nil && wlIDs == nil && blMap == nil && wlMap == nil {
		return nil
	}
	return &simplePublishFilter{blIDs, wlIDs, blMap, wlMap, len(blMap) != 0 || len(wlMap) != 0}
}

// LockRequired determines whether a consistent state of the subscriber sessions is
// required while running the filter
func (f *simplePublishFilter) LockRequired() bool {
	return f.lockRequired
}

// PublishAllowed determines if a message is allowed to be published to a
// subscriber, by looking at any blacklists and whitelists provided with the
// publish message.
//
// To receive a published event, the subscriber session must not have any
// values that appear in a blacklist, and must have a value from each
// whitelist.
func (f *simplePublishFilter) PublishAllowed(sub *wamp.Session) bool {
	// Check each blacklisted ID to see if session ID is blacklisted.
	for i := range f.blIDs {
		if f.blIDs[i] == sub.ID {
			return false
		}
	}

	var eligible bool
	// If session ID whitelist given, make sure session ID is in whitelist.
	if len(f.wlIDs) != 0 {
		for i := range f.wlIDs {
			if f.wlIDs[i] == sub.ID {
				eligible = true
				break
			}
		}
		if !eligible {
			return false
		}
	}

	// Check blacklists to see if session has a value in any blacklist.
	for attr, vals := range f.blMap {
		// Get the session attribute value to compare with blacklist.
		sessAttr, _ := wamp.AsString(sub.Details[attr])
		if sessAttr == "" {
			continue
		}
		// Check each blacklisted value to see if session attribute is one.
		for i := range vals {
			if vals[i] == sessAttr {
				// Session has blacklisted attribute value.
				return false
			}
		}
	}

	// Check whitelists to make sure session has value in each whitelist.
	for attr, vals := range f.wlMap {
		// Get the session attribute value to compare with whitelist.
		sessAttr, _ := wamp.AsString(sub.Details[attr])
		if sessAttr == "" {
			// Session does not have whitelisted value, so deny.
			return false
		}
		eligible = false
		// Check all whitelisted values to see is session attribute is one.
		for i := range vals {
			if vals[i] == sessAttr {
				// Session has whitelisted attribute value.
				eligible = true
				break
			}
		}
		// If session attribute value no found in whitelist, then deny.
		if !eligible {
			return false
		}
	}
	return true
}
