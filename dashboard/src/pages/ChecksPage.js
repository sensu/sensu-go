import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";

import Paper from "material-ui/Paper";
import AppContent from "../components/AppContent";
import CheckList from "../components/CheckList";

class CheckPage extends React.Component {
  static propTypes = {
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
  };

  static query = graphql`
    query ChecksPageQuery {
      viewer {
        ...CheckList_viewer
      }
    }
  `;

  render() {
    const { viewer } = this.props;
    return (
      <AppContent>
        <Paper>
          <CheckList viewer={viewer} />
        </Paper>
      </AppContent>
    );
  }
}

export default CheckPage;
