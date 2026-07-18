export const SLUG_MAX = 40;
export const SLUG_RE = /^[a-z][a-z0-9-]{0,39}$/;

function trimDashes(value: string): string {
	return value.replace(/^-+/, '').replace(/-+$/, '');
}

export function slugify(name: string): string {
	let out = '';
	let lastDash = false;
	for (const ch of name.trim().toLowerCase()) {
		if ((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			out += ch;
			lastDash = false;
		} else if (ch === ' ' || ch === '-' || ch === '_') {
			if (!lastDash && out.length > 0) {
				out += '-';
				lastDash = true;
			}
		}
	}
	out = trimDashes(out);
	if (out.length > SLUG_MAX) out = trimDashes(out.slice(0, SLUG_MAX));
	if (out === '' || out[0] < 'a' || out[0] > 'z') out = 'w' + out;
	return out;
}
