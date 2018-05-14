import React from "react";
import PropTypes from "prop-types";
import { Form } from "@10xjs/form";

import TargetsPanel from "./SilenceEntryFormTargetsPanel";
import CheckPanel from "./SilenceEntryFormCheckPanel";
import SubscriptionPanel from "./SilenceEntryFormSubscriptionPanel";
import ExpirationPanel from "./SilenceEntryFormExpirationPanel";
import SchedulePanel from "./SilenceEntryFormSchedulePanel";
import ReasonPanel from "./SilenceEntryFormReasonPanel";

class SilenceEntryForm extends React.PureComponent {
  static propTypes = {
    onSubmit: PropTypes.func,
    onSubmitSuccess: PropTypes.func,
    values: PropTypes.object,
  };

  static defaultProps = {
    onSubmit: undefined,
    onSubmitSuccess: undefined,
    values: undefined,
  };

  formRef = React.createRef();

  submit() {
    this.formRef.current.submit();
  }

  render() {
    const { values, onSubmit, onSubmitSuccess } = this.props;

    return (
      <Form
        ref={this.formRef}
        onSubmit={onSubmit}
        onSubmitSuccess={onSubmitSuccess}
        values={values}
      >
        {({ submit, setValue }) => (
          <form onSubmit={submit}>
            {Array.isArray(values.targets) ? (
              <TargetsPanel />
            ) : (
              <React.Fragment>
                <CheckPanel />
                <SubscriptionPanel />
              </React.Fragment>
            )}
            <SchedulePanel />
            <ExpirationPanel setFieldValue={setValue} />
            <ReasonPanel />
          </form>
        )}
      </Form>
    );
  }
}
export default SilenceEntryForm;
