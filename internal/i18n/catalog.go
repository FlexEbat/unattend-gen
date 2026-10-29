package i18n

// catalog maps an English UI text to its translations (ru, zh, es, hi).
var catalog = map[string][4]string{}

// msgCatalog maps a canonical Russian validation-message template to its
// translations (en, zh, es, hi).
var msgCatalog = map[string][4]string{}

// keep lists UI texts that are identical in every language (technical
// tokens, file extensions, placeholders); they need no translation.
var keep = map[string]struct{}{}

// add registers UI translations: key (English), ru, zh, es, hi.
func add(rows ...[5]string) {
	for _, r := range rows {
		if _, dup := catalog[r[0]]; dup {
			panic("i18n: duplicate UI key " + r[0])
		}
		catalog[r[0]] = [4]string{r[1], r[2], r[3], r[4]}
	}
}

// addMsg registers message translations: key (Russian template), en, zh, es, hi.
func addMsg(rows ...[5]string) {
	for _, r := range rows {
		if _, dup := msgCatalog[r[0]]; dup {
			panic("i18n: duplicate message key " + r[0])
		}
		msgCatalog[r[0]] = [4]string{r[1], r[2], r[3], r[4]}
	}
}

// same marks UI texts that stay as they are in every language.
func same(keys ...string) {
	for _, k := range keys {
		keep[k] = struct{}{}
	}
}
