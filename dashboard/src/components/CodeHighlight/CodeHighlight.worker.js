import hljs from "highlight.js/lib/highlight";
import bash from "highlight.js/lib/languages/bash";
import properties from "highlight.js/lib/languages/properties";

hljs.registerLanguage("bash", bash);
hljs.registerLanguage("properties", properties);

onmessage = message => {
  const [language, data] = message.data;
  const result = hljs.highlight(language, data);
  postMessage(result.value);
};
