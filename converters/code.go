package converters

import "strings"

// FormatJavaCodeStr will remove from the provided java code new lines and white space's
// Input example:
/*`
	if (!ctx._source.containsKey('roles')) {
		ctx._source.roles = new HashMap();
	}
	if (!ctx._source.roles.containsKey(params.role)) {
		ctx._source.roles.put(params.role, [params.address]);
	} else {
		int i;
		for (i = 0; i < ctx._source.roles.get(params.role).length; i++) {
			if (ctx._source.roles.get(params.role).get(i) == params.address) {
				return;
			}
		}
		ctx._source.roles.get(params.role).add(params.address);
	}
` */
// OUTPUT: `if (!ctx._source.containsKey('roles')) {ctx._source.roles = new HashMap();}if (!ctx._source.roles.containsKey(params.role)) {ctx._source.roles.put(params.role, [params.address]);} else {int i;for (i = 0; i < ctx._source.roles.get(params.role).length; i++) {if (ctx._source.roles.get(params.role).get(i) == params.address) {return;}}ctx._source.roles.get(params.role).add(params.address);}`
func FormatJavaCodeStr(code string) string {
	formatted := strings.ReplaceAll(code, "\n", "")
	formatted = strings.ReplaceAll(formatted, "\t", "")

	return formatted
}
