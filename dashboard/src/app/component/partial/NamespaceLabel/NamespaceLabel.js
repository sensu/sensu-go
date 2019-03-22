import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Label from "./NamespaceLabelBase";

class NamespaceLabel extends React.Component {
  static propTypes = {
    namespace: PropTypes.shape({
      name: PropTypes.string,
      colourId: PropTypes.string,
      iconId: PropTypes.string,
    }).isRequired,
  };

  static fragments = {
    namespace: gql`
      fragment NamespaceLabel_namespace on Namespace {
        name
        colourId
        iconId
      }
    `,
  };

  render() {
    const { namespace, ...props } = this.props;

    return (
      <Label
        name={namespace.name}
        icon={namespace.iconId}
        colour={namespace.colourId}
        {...props}
      />
    );
  }
}

export default NamespaceLabel;
