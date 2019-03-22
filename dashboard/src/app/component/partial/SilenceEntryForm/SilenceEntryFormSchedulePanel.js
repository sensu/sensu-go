import React from "react";
import { Field } from "@10xjs/form";
import DateInputController from "@10xjs/date-input-controller";

import { withStyles } from "@material-ui/core/styles";
import Collapse from "@material-ui/core/Collapse";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Switch from "@material-ui/core/Switch";
import Typography from "@material-ui/core/Typography";
import Select from "@material-ui/core/Select";
import InputLabel from "@material-ui/core/InputLabel";

import {
  getMonthName,
  getTimeZoneName,
  getHour,
  getDayperiod,
} from "/lib/util/date";

import DateFormatter from "/lib/component/base/DateFormatter";

import Panel from "./SilenceEntryFormPanel";

const StyledFormControl = withStyles(theme => ({
  root: {
    marginTop: theme.spacing.unit,
    marginBottom: theme.spacing.unit,
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
    "& + &": {
      marginLeft: 0,
    },
  },
  fullWidth: {
    marginLeft: 0,
    marginRight: 0,
  },
}))(FormControl);

const scopeId = id => `SilenceEntryFormSchedulePanel-${id}`;

const range = (start, end) =>
  Array(...Array(1 + end - start)).map((_, i) => start + i);

// eslint-disable-next-line react/no-multi-comp
class SilenceEntryFormSchedulePanel extends React.PureComponent {
  constructor(props) {
    super(props);

    this._currentDate = new Date();

    this._minBeginDate = new Date();
    this._minBeginDate.setMinutes(0);
    this._minBeginDate.setSeconds(0);

    this._maxBeginDate = new Date();
    this._maxBeginDate.setFullYear(this._maxBeginDate.getFullYear() + 10);
    this._maxBeginDate.setMinutes(0);
    this._maxBeginDate.setSeconds(0);

    this._lastBeginValue = new Date();
    this._lastBeginValue.setHours(this._lastBeginValue.getHours() + 1);
    this._lastBeginValue.setMinutes(0);
    this._lastBeginValue.setSeconds(0);
  }

  render() {
    return (
      <Field
        path="props.begin"
        format={value =>
          value === null || value === undefined ? null : new Date(value)
        }
        parse={value =>
          value === null || value === undefined ? null : value.toISOString()
        }
      >
        {({ setValue, input, composeInput }) => (
          <Panel
            title={"Schedule"}
            summary={
              input.value === null ? (
                "begin immediately"
              ) : (
                <React.Fragment>
                  begin on{" "}
                  <DateFormatter
                    month="short"
                    day="numeric"
                    year={
                      input.value.getFullYear() !==
                      this._currentDate.getFullYear()
                        ? "numeric"
                        : undefined
                    }
                    value={input.value}
                  />{" "}
                  at{" "}
                  <DateFormatter
                    hour="numeric"
                    minute="numeric"
                    timeZoneName="short"
                    value={input.value}
                  />
                </React.Fragment>
              )
            }
            hasDefaultValue={input.value === null}
          >
            <Typography color="textSecondary">
              Silencing will begin immediately when the entry is created, or can
              be scheduled to begin at a later date and time.
            </Typography>

            <StyledFormControl fullWidth>
              <FormControlLabel
                control={
                  <Switch
                    checked={input.value !== null}
                    onChange={event => {
                      const checked = event.target.checked;
                      setValue(checked ? this._lastBeginValue : null);
                    }}
                  />
                }
                label="Schedule silencing at a later date and time"
              />
            </StyledFormControl>
            <Collapse in={input.value !== null}>
              <DateInputController
                min={this._minBeginDate}
                max={this._maxBeginDate}
                {...composeInput({
                  onChange: value => {
                    this._lastBeginValue = value;
                  },
                })}
                value={input.value || this._lastBeginValue}
              >
                {date => (
                  <Typography component="div">
                    <div>
                      on{" "}
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("year")}>Year</InputLabel>
                        <Select
                          native
                          value={date.year}
                          onChange={event => date.setYear(event.target.value)}
                          inputProps={{
                            name: "year",
                            id: scopeId("year"),
                          }}
                        >
                          {range(date.yearMin, date.yearMax).map(value => (
                            <option key={value} value={value}>
                              {value}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("month")}>
                          Month
                        </InputLabel>
                        <Select
                          native
                          value={date.month}
                          onChange={event => date.setMonth(event.target.value)}
                          inputProps={{
                            name: "month",
                            id: scopeId("month"),
                          }}
                        >
                          {range(date.monthMin, date.monthMax).map(value => (
                            <option key={value} value={value}>
                              {getMonthName(new Date(1970, value, 1))}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("day")}>Day</InputLabel>
                        <Select
                          native
                          value={date.day}
                          onChange={event => date.setDay(event.target.value)}
                          inputProps={{
                            name: "day",
                            id: scopeId("day"),
                          }}
                        >
                          {range(date.dayMin, date.dayMax).map(value => (
                            <option key={value} value={value}>
                              {value}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>{" "}
                    </div>
                    <div>
                      at{" "}
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("hour")}>Hour</InputLabel>
                        <Select
                          native
                          value={date.hour}
                          onChange={event => date.setHour(event.target.value)}
                          inputProps={{
                            name: "hour",
                            id: scopeId("hour"),
                          }}
                        >
                          {range(date.hourMin, date.hourMax).map(value => (
                            <option key={value} value={value}>
                              {Number(
                                getHour(new Date(1970, 0, 1, value)),
                              ).toLocaleString("en-US", {
                                minimumIntegerDigits: 2,
                              })}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>
                      :
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("minute")}>
                          Minute
                        </InputLabel>
                        <Select
                          native
                          value={date.minute}
                          onChange={event => date.setMinute(event.target.value)}
                          inputProps={{
                            name: "minute",
                            id: scopeId("minute"),
                          }}
                        >
                          {range(date.minuteMin, date.minuteMax).map(value => (
                            <option key={value} value={value}>
                              {value.toLocaleString("en-US", {
                                minimumIntegerDigits: 2,
                              })}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>
                      :
                      <StyledFormControl>
                        <InputLabel htmlFor={scopeId("second")}>
                          Second
                        </InputLabel>
                        <Select
                          native
                          value={date.second}
                          onChange={event => date.setSecond(event.target.value)}
                          inputProps={{
                            name: "second",
                            id: scopeId("second"),
                          }}
                        >
                          {range(date.secondMin, date.secondMax).map(value => (
                            <option key={value} value={value}>
                              {value.toLocaleString("en-US", {
                                minimumIntegerDigits: 2,
                              })}
                            </option>
                          ))}
                        </Select>
                      </StyledFormControl>{" "}
                      {getDayperiod(input.value)} {getTimeZoneName(input.value)}
                    </div>
                  </Typography>
                )}
              </DateInputController>
            </Collapse>
          </Panel>
        )}
      </Field>
    );
  }
}

export default SilenceEntryFormSchedulePanel;
