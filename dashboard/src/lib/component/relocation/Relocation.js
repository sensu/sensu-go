/* eslint-disable react/no-multi-comp */
/* eslint-disable react/sort-comp */
/* eslint-disable react/no-unused-state */

import React from "react";
import PropTypes from "prop-types";

import uniqueId from "/lib/util/uniqueId";

const mergeAtIndex = (arr, index, update) =>
  arr
    .slice(0, index)
    .concat([
      {
        ...arr[index],
        ...update,
      },
    ])
    .concat(arr.slice(index + 1));

const removeAtIndex = (arr, index) =>
  arr.slice(0, index).concat(arr.slice(index + 1));

const Context = React.createContext({
  addChild: () => {},
  updateChild: () => {},
  removeChild: () => {},
  elements: [],
});

class SinkConnector extends React.PureComponent {
  static propTypes = {
    addChild: PropTypes.func.isRequired,
    updateChild: PropTypes.func.isRequired,
    removeChild: PropTypes.func.isRequired,
    children: PropTypes.any.isRequired,
  };

  id = uniqueId();

  componentDidMount() {
    this.props.addChild(this.id, this.props.children);
  }

  componentDidUpdate() {
    this.props.updateChild(this.id, this.props.children);
  }

  componentWillUnmount() {
    this.props.removeChild(this.id);
  }

  render() {
    return null;
  }
}

class ConsumerRender extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    const { children, ...props } = this.props;
    return children(props);
  }
}

export class Provider extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node,
  };

  static defaultProps = { children: undefined };

  createChild = props => {
    this.setState(state => {
      const id = uniqueId();

      const element = {
        update: nextProps => this.updateChild(id, nextProps),
        remove: () => this.removeChild(id),
        id,
        props,
      };

      const elements = state.elements.concat([element]);

      return { elements };
    });
  };

  addChild = (id, props) => {
    this.setState(state => {
      const index = state.elements.findIndex(element => element.id === id);

      if (index !== -1) {
        throw new Error("Duplicate relocation child ID");
      }

      const element = { id, props };

      const elements = state.elements.concat([element]);

      return { elements };
    });
  };

  updateChild = (id, props) => {
    this.setState(state => {
      const index = state.elements.findIndex(element => element.id === id);

      if (index === -1) {
        throw new Error("Invalid relocation child ID");
      }

      const elements = mergeAtIndex(state.elements, index, { props });

      return { elements };
    });
  };

  removeChild = id => {
    this.setState(state => {
      const index = state.elements.findIndex(element => element.id === id);

      if (index === -1) {
        return null;
      }

      const elements = removeAtIndex(state.elements, index);

      return { elements };
    });
  };

  state = {
    createChild: this.createChild,
    addChild: this.addChild,
    updateChild: this.updateChild,
    removeChild: this.removeChild,
    elements: [],
  };

  render() {
    return (
      <Context.Provider value={this.state}>
        {this.props.children}
      </Context.Provider>
    );
  }
}

export class Sink extends React.PureComponent {
  static propTypes = {
    children: PropTypes.any.isRequired,
  };

  render() {
    return (
      <Context.Consumer>
        {({ addChild, updateChild, removeChild }) => (
          <SinkConnector
            addChild={addChild}
            updateChild={updateChild}
            removeChild={removeChild}
          >
            {this.props.children}
          </SinkConnector>
        )}
      </Context.Consumer>
    );
  }
}

export class Well extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return (
      <Context.Consumer>
        {({ elements }) => (
          <ConsumerRender elements={elements}>
            {this.props.children}
          </ConsumerRender>
        )}
      </Context.Consumer>
    );
  }
}

export class Consumer extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return (
      <Context.Consumer>
        {({ createChild, removeChild }) => (
          <ConsumerRender createChild={createChild} removeChild={removeChild}>
            {this.props.children}
          </ConsumerRender>
        )}
      </Context.Consumer>
    );
  }
}
