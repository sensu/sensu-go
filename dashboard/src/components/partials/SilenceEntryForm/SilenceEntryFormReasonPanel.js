import React from "react";
import { Field } from "@10xjs/form";

import Typography from "@material-ui/core/Typography";
import TextField from "@material-ui/core/TextField";

import Panel from "./SilenceEntryFormPanel";

class SilenceEntryFormReasonPanel extends React.PureComponent {
  render() {
    return (
      <Field path="reason">
        {reason => (
          <Panel
            title="Reason"
            summary={reason.props.value}
            hasDefaultValue={!reason.rawValue}
          >
            <Typography color="textSecondary">
              Explanation for the creation of this entry.
            </Typography>

            <TextField
              label="Reason"
              multiline
              fullWidth
              rowsMax="4"
              margin="normal"
              {...reason.props}
            />
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormReasonPanel;
