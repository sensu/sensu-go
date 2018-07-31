import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { compose } from "recompose";
import { withRouter } from "react-router-dom";
import { withApollo } from "react-apollo";

import executeCheck from "/mutations/executeCheck";

class CheckDetailsExecuteAction extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    check: PropTypes.object,
  };

  static defaultProps = {
    check: null,
  };

  static fragments = {
    check: gql`
      fragment CheckDetailsExecuteAction_check on CheckConfig {
        id
      }
    `,
  };

  handleClick = () => {
    const { client, check } = this.props;
    executeCheck(client, { id: check.id });
  };

  render() {
    return this.props.children(this.handleClick);
  }
}

const enhancer = compose(withApollo, withRouter);
export default enhancer(CheckDetailsExecuteAction);
