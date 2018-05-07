import React from "react";
import { Field } from "@10xjs/form";
import { withStyles } from "material-ui/styles";

import Typography from "material-ui/Typography";
import TextField from "material-ui/TextField";

import Panel from "./SilenceEntryFormPanel";

const MonoTextField = withStyles(theme => ({
  root: { "& input": { fontFamily: theme.typography.monospace.fontFamily } },
}))(TextField);

class SilenceEntryFormCheckPanel extends React.PureComponent {
  render() {
    return (
      <Field path="check">
        {check => (
          <Panel
            title="Check"
            summary={check.props.value || "all checks"}
            hasDefaultValue={!check.stateValue}
          >
            <Typography color="textSecondary">
              Enter the name of a check the silencing entry should match.
            </Typography>

            <MonoTextField
              label="Check"
              fullWidth
              margin="normal"
              {...check.props}
            />
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormCheckPanel;
