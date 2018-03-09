import React from "react";
import PropTypes from "prop-types";

import { graphql } from "react-relay";
import Typography from "material-ui/Typography";
import { withStyles } from "material-ui/styles";

import AppContent from "../components/AppContent";
import EventsContainer from "../components/EventsContainer";
import SearchBox from "../components/SearchBox";

const styles = {
  headline: {
    display: "flex",
    justifyContent: "space-between",
    alignContent: "center",
  },
  title: {
    display: "flex",
    alignSelf: "flex-end",
  },
  container: {
    marginTop: 10,
  },
};

class EventsPage extends React.Component {
  static propTypes = {
    // eslint-disable-next-line react/forbid-prop-types
    classes: PropTypes.object.isRequired,
  };

  static query = graphql`
    query EventsPageQuery(
      $filter: String
      $order: EventsListOrder = SEVERITY
      $environment: String!
      $organization: String!
    ) {
      environment(organization: $organization, environment: $environment) {
        ...EventsContainer_environment
      }
    }
  `;

  state = { inputValue: "" };

  requerySearchBox = inputValue => {
    this.setState({ inputValue });
    // TODO return to this and make it actually query
    // eslint-disable-next-line no-console
    console.info("inputValue", inputValue);
  };

  render() {
    const { classes, ...props } = this.props;
    return (
      <AppContent>
        <div>
          <div className={classes.headline}>
            <Typography className={classes.title} type="headline">
              Recent Events
            </Typography>
            <SearchBox
              onUpdateInput={this.requerySearchBox}
              state={this.state.inputValue}
            />
          </div>
          <div className={classes.container}>
            <EventsContainer {...props} />
          </div>
        </div>
      </AppContent>
    );
  }
}

export default withStyles(styles)(EventsPage);
