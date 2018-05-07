import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { withField } from "@10xjs/form";

import { FormControl, FormControlLabel } from "material-ui/Form";
import TextField from "material-ui/TextField";
import Switch from "material-ui/Switch";
import { InputAdornment } from "material-ui/Input";
import Typography from "material-ui/Typography";

import Panel from "./SilenceEntryFormPanel";

const DEFAULT_EXPIRE_DURATION = 3600;

const parseNumber = value => {
  const number = parseInt(value, 10);
  return isNaN(number) ? -1 : number;
};

const formatNumber = value => (value === -1 ? "" : `${value}`);

class SilenceEntryFormExpirationPanel extends React.PureComponent {
  static propTypes = {
    expireOnResolve: PropTypes.object.isRequired,
    expire: PropTypes.object.isRequired,
    setFieldValue: PropTypes.func.isRequired,
  };

  render() {
    const { expireOnResolve, expire, setFieldValue } = this.props;

    const expireAfterDuration = expire.stateValue > 0;

    if (expireAfterDuration) {
      this._lastExprireValue = expire.props.value;
    }

    const hasDefaultValue =
      !expireOnResolve.props.checked && !expireAfterDuration;

    const summary =
      [
        expireOnResolve.props.checked ? "on resolved check" : null,
        expireAfterDuration
          ? `after ${expire.stateValue} ${
              expire.stateValue === 1 ? "second" : "seconds"
            }`
          : null,
      ]
        .filter(Boolean)
        .join(", or ") || "when manually removed";

    return (
      <Panel
        title="Expiration"
        summary={summary}
        hasDefaultValue={hasDefaultValue}
      >
        <Typography color="textSecondary">
          This silencing entry will be automatically removed when any of the
          expiration conditions are met.
        </Typography>
        <FormControl fullWidth>
          <FormControlLabel
            control={<Switch {...expireOnResolve.props} />}
            label="Expire when a matching check resolves"
          />
        </FormControl>
        <FormControl fullWidth>
          <FormControlLabel
            control={
              <Switch
                checked={expireAfterDuration}
                onChange={event => {
                  const checked = event.target.checked;
                  setFieldValue(
                    "expire",
                    checked
                      ? this._lastExprireValue || DEFAULT_EXPIRE_DURATION
                      : -1,
                  );
                }}
              />
            }
            label="Expire after a fixed duration"
          />
        </FormControl>
        {// WIP: react-motion react-resize-observer expander thing here
        (expireAfterDuration || expire.focused) && (
          <FormControl fullWidth>
            <TextField
              type="number"
              label="Expire after"
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">seconds</InputAdornment>
                ),
              }}
              {...expire.composeProps({
                onChange: event => {
                  setFieldValue("check", Math.random());
                  if (!event.target.value) {
                    this._lastExprireValue = undefined;
                    setFieldValue("expire", -1);
                  }
                },
              })}
            />
          </FormControl>
        )}
      </Panel>
    );
  }
}

export default compose(
  withField("expireOnResolve", { path: "expireOnResolve", checkbox: true }),
  withField("expire", {
    path: "expire",
    parse: parseNumber,
    format: formatNumber,
  }),
)(SilenceEntryFormExpirationPanel);
