import React from "react";
import { Field } from "@10xjs/form";

import {
  UNIQUE_CONSTRAINT,
  REQUIRED,
  formatValidationError,
} from "/utils/validation";

import TargetsPanel from "./SilenceEntryFormTargetsPanel";
import CheckPanel from "./SilenceEntryFormCheckPanel";
import SubscriptionPanel from "./SilenceEntryFormSubscriptionPanel";
import ExpirationPanel from "./SilenceEntryFormExpirationPanel";
import SchedulePanel from "./SilenceEntryFormSchedulePanel";
import ReasonPanel from "./SilenceEntryFormReasonPanel";

const formatCheckSubscriptionError = error => {
  if (!error) {
    return "";
  }

  return (
    {
      [REQUIRED]: "Either check or subscription is required.",
      [UNIQUE_CONSTRAINT]: "Cannot create duplicate silencing entry.",
    }[error.code] || formatValidationError(error)
  );
};

class SilenceEntryFormFields extends React.PureComponent {
  render() {
    return (
      <React.Fragment>
        <Field path="targets">
          {({ rawValue }) =>
            Array.isArray(rawValue) ? (
              <TargetsPanel formatError={formatCheckSubscriptionError} />
            ) : (
              <React.Fragment>
                <CheckPanel formatError={formatCheckSubscriptionError} />
                <SubscriptionPanel formatError={formatCheckSubscriptionError} />
              </React.Fragment>
            )
          }
        </Field>
        <SchedulePanel />
        <ExpirationPanel />
        <ReasonPanel />
      </React.Fragment>
    );
  }
}

export default SilenceEntryFormFields;
