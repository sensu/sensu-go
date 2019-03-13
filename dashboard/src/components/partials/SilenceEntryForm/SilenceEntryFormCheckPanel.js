import React from "react";
import PropTypes from "prop-types";
import { Field } from "@10xjs/form";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import TextField from "@material-ui/core/TextField";

import ResetAdornment from "/components/partials/ResetAdornment";

import Panel from "./SilenceEntryFormPanel";

const MonoTextField = withStyles(theme => ({
  root: { "& input": { fontFamily: theme.typography.monospace.fontFamily } },
}))(TextField);

class SilenceEntryFormCheckPanel extends React.PureComponent {
  static propTypes = {
    formatError: PropTypes.func.isRequired,
  };

  render() {
    const { formatError } = this.props;

    return (
      <Field path="check">
        {({ input, rawValue, error, dirty, initialValue, setValue }) => (
          <Panel
            title="Check"
            summary={input.value || "all checks"}
            hasDefaultValue={!rawValue}
            error={formatError(error)}
          >
            <Typography color="textSecondary">
              Enter the name of a check the silencing entry should match.
            </Typography>

            <MonoTextField
              label="Check"
              fullWidth
              margin="normal"
              error={!!formatError(error)}
              InputProps={{
                endAdornment: initialValue && dirty && (
                  <ResetAdornment onClick={() => setValue(initialValue)} />
                ),
              }}
              {...input}
            />
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormCheckPanel;
