import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";

import Paper from "material-ui/Paper";
import AppContent from "../components/AppContent";
import EventList from "../components/EventList";

class EventsPage extends React.Component {
  static propTypes = {
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
  };

  static query = graphql`
    query EventsPageQuery {
      viewer {
        ...EventList_viewer
      }
    }
  `;

  render() {
    const { viewer } = this.props;
    return (
      <AppContent>
        <Paper>
          <EventList viewer={viewer} />
        </Paper>
      </AppContent>
    );
  }
}

export default EventsPage;
