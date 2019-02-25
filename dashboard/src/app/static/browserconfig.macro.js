import icon from "./favicon-180.png";

export default ({ emitFile }) =>
  emitFile(
    "browserconfig.xml",
    `
<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
  <msapplication>
    <tile>
      <square70x70logo src="${icon}"/>
      <square150x150logo src=""${icon}"/>
      <TileColor>#2D3555</TileColor>
    </tile>
  </msapplication>
</browserconfig>
`.slice(1),
  );
