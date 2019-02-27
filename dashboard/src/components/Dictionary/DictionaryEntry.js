import React from "react";
import PropTypes from "prop-types";
import warning from "warning";

const Context = React.createContext();

class DictionaryEntry extends React.Component {
  static propTypes = {
    fullWidth: PropTypes.bool,
    children: PropTypes.node.isRequired,
  };

  static defaultProps = {
    fullWidth: false,
  };

  static Context = Context.Consumer;

  render() {
    const { children: childrenProp, fullWidth, ...props } = this.props;
    warning(
      childrenProp.length === 2,
      "DictionaryEntry component should onle recieve only two children a key and a value.",
    );
    let children;
    if (fullWidth) {
      children = React.Children.map(childrenProp, child => (
        <tr {...props}>{child}</tr>
      ));
    } else {
      children = <tr {...props}>{childrenProp}</tr>;
    }
    return (
      <Context.Provider value={{ fullWidth }}>{children}</Context.Provider>
    );
  }
}

export default DictionaryEntry;
