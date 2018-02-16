import React from "react";
import PropTypes from "prop-types";
import { graphql } from "react-relay";
import Typography from "material-ui/Typography";

import AppContent from "../components/AppContent";
import EventsContainer from "../components/EventsContainer";

class EventsPage extends React.Component {
  static propTypes = {
    viewer: PropTypes.objectOf(PropTypes.any).isRequired,
  };

  static query = graphql`
    query EventsPageQuery($filter: String) {
      viewer {
        ...EventsContainer_viewer
      }
    }
  `;

  render() {
    const { viewer } = this.props;
    return (
      <AppContent>
        <div>
          <Typography type="headline">Recent Events</Typography>
          <EventsContainer viewer={viewer} />
        </div>
      </AppContent>
    );
  }
}

export default EventsPage;
