import React from "react";
import { Field } from "@10xjs/form";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import TextField from "@material-ui/core/TextField";

import Code from "/components/Code";

import Panel from "./SilenceEntryFormPanel";

const MonoTextField = withStyles(theme => ({
  root: { "& input": { fontFamily: theme.typography.monospace.fontFamily } },
}))(TextField);

class SilenceEntryFormSubscriptionPanel extends React.PureComponent {
  render() {
    return (
      <Field path="subscription">
        {subscription => (
          <Panel
            title="Subscription"
            summary={subscription.props.value || "all entities"}
            hasDefaultValue={!subscription.stateValue}
          >
            <Typography color="textSecondary">
              Enter the name of the subscription the entry should match. Use the
              format <Code>entity:$ENTITY_NAME</Code> to match a specific
              entity.
            </Typography>

            <MonoTextField
              label="Subscription"
              fullWidth
              margin="normal"
              {...subscription.props}
            />
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormSubscriptionPanel;
