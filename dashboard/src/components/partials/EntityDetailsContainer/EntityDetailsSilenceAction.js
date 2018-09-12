import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import SilenceEntryDialog from "/components/partials/SilenceEntryDialog";

class EntityDetailsSilenceAction extends React.Component {
  static propTypes = {
    children: PropTypes.func.isRequired,
    entity: PropTypes.object.isRequired,
    onDone: PropTypes.func,
  };

  static defaultProps = {
    onDone: () => false,
  };

  static fragments = {
    entity: gql`
      fragment EntityDetailsSilenceAction_entity on Entity {
        namespace {
          organization
          environment
        }

        name
        isSilenced
      }
    `,
  };

  state = { isOpen: false };

  render() {
    const { entity } = this.props;
    const { isOpen } = this.state;

    const canOpen = entity.isSilenced;
    const open = () => this.setState({ isOpen: true });

    return (
      <React.Fragment>
        {this.props.children({ canOpen, open })}
        {isOpen && (
          <SilenceEntryDialog
            values={{
              check: "*",
              subscription: `entity:${entity.name}`,
              ns: {
                organization: entity.namespace.organization,
                environment: entity.namespace.environment,
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

export default EntityDetailsSilenceAction;
