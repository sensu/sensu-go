import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import { withField } from "@10xjs/form";

import Collapse from "@material-ui/core/Collapse";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import TextField from "@material-ui/core/TextField";
import Switch from "@material-ui/core/Switch";
import InputAdornment from "@material-ui/core/InputAdornment";
import Typography from "@material-ui/core/Typography";

import Panel from "./SilenceEntryFormPanel";

const DEFAULT_EXPIRE_DURATION = 3600;

class SilenceEntryFormExpirationPanel extends React.PureComponent {
  static propTypes = {
    expireOnResolve: PropTypes.object.isRequired,
    expire: PropTypes.object.isRequired,
  };

  render() {
    const { expireOnResolve, expire } = this.props;

    const expireAfterDuration = expire.rawValue > 0;

    if (expireAfterDuration) {
      this._lastExprireValue = expire.input.value;
    }

    const hasDefaultValue =
      !expireOnResolve.input.checked && !expireAfterDuration;

    const summary =
      [
        expireOnResolve.input.checked ? "on resolved check" : null,
        expireAfterDuration
          ? `after ${expire.rawValue} ${
              expire.rawValue === 1 ? "second" : "seconds"
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
            control={<Switch {...expireOnResolve.input} />}
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
                  expire.setValue(
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
        <Collapse in={expireAfterDuration || expire.focused}>
          <FormControl fullWidth>
            <TextField
              type="number"
              label="Expire after"
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">seconds</InputAdornment>
                ),
              }}
              {...expire.composeInput({
                onChange: event => {
                  if (!event.target.value) {
                    this._lastExprireValue = undefined;
                    expire.setValue(-1);
                  }
                },
              })}
            />
          </FormControl>
        </Collapse>
      </Panel>
    );
  }
}

export default compose(
  withField("expireOnResolve", {
    path: "props.expireOnResolve",
    checkbox: true,
  }),
  withField("expire", {
    path: "props.expire",
    parse(value) {
      const number = parseInt(value, 10);
      return isNaN(number) ? -1 : number;
    },
    format(value) {
      if (value === undefined || value === null || value === -1) {
        return "";
      }
      return `${value}`;
    },
  }),
)(SilenceEntryFormExpirationPanel);
