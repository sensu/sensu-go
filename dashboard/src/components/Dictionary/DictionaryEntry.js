import React from "react";
import PropTypes from "prop-types";
import warning from "warning";

class DictionaryEntry extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
  };

  render() {
    const { children, ...props } = this.props;
    warning(
      children.length === 2,
      "DictionaryEntry component should onle recieve only two children a key and a value.",
    );
    return <tr {...props}>{children}</tr>;
  }
}

export default DictionaryEntry;
