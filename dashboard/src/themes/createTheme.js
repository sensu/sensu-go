import deepmerge from "deepmerge";
import createMuiTheme from "material-ui/styles/createMuiTheme";
import createPalette from "material-ui/styles/createPalette";
import createTypography from "material-ui/styles/createTypography";

import defaults from "./defaults";

function getTypography(typography, palette) {
  if (typeof typography === "function") {
    return typography(palette);
  }
  return typography;
}

function createTheme(theme) {
  // Apply palette first so that we can use it else where
  const paletteInputs = deepmerge(defaults.palette, theme.palette);
  const palette = createPalette(paletteInputs);

  // Create typography using defaults and configured palette
  const typographyInputs = deepmerge(
    getTypography(defaults.typography, palette),
    getTypography(theme.typography || {}, palette),
  );
  const typography = createTypography(palette, typographyInputs);

  // Merge remaining attributes
  const rest = deepmerge(defaults, theme);
  return createMuiTheme({
    ...rest,
    palette,
    typography,
  });
}

export default createTheme;
