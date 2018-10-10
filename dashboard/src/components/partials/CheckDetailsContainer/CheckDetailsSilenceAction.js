import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";

class CheckDetailsSilenceAction extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    check: PropTypes.object.isRequired,
    onDone: PropTypes.func,
  };

  static defaultProps = {
    onDone: () => false,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsSilenceAction_check on CheckConfig {
        name
        namespace {
          organization
          environment
        }

        isSilenced
      }
    `,
  };

  state = { isOpen: false };

  render() {
    const { check } = this.props;
    const { isOpen } = this.state;

    const canOpen = check.isSilenced;
    const open = () => this.setState({ isOpen: true });

    return (
      <React.Fragment>
        {this.props.children({ canOpen, open })}
        {isOpen && (
          <SilenceEntryDialog
            values={{
              check: check.name,
              subscription: "*",
              ns: {
                organization: check.namespace.organization,
                environment: check.namespace.environment,
              },
            }}
            onClose={() => {
              this.props.onDone();
              this.setState({ isOpen: false });
            }}
          />
        )}
      </React.Fragment>
    );
  }
}

export default CheckDetailsSilenceAction;
