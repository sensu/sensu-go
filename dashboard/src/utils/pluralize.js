import { plural } from "pluralize";

function shouldBePlural(n, { locale = "en-us" }) {
  const rules = new Intl.PluralRules(locale);
  const rule = rules.select(n);
  return rule !== "one";
}

function toPlural(word, n = 2, opts = {}) {
  if (shouldBePlural(n, opts)) {
    return plural(word);
  }
  return word;
}

export { shouldBePlural, toPlural };
