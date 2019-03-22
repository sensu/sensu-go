import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import SilenceEntryDialog from "/app/component/partial/SilenceEntryDialog";

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
        name
        namespace
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
              namespace: entity.namespace,
              subscription: `entity:${entity.name}`,
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
