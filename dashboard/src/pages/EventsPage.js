import React from "react";
import PropTypes from "prop-types";

import { graphql } from "react-relay";
import { routerShape, matchShape } from "found";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";

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
    classes: PropTypes.object.isRequired,
    location: PropTypes.shape({
      query: PropTypes.object.isRequired,
    }).isRequired,
    router: routerShape.isRequired,
    match: matchShape.isRequired,
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

  constructor(props) {
    super(props);
    this.state = {
      filterValue: props.location.query.filter,
    };
  }

  changeQuery = (key, val) => {
    const { match, router } = this.props;
    const query = new URLSearchParams(match.location.query);

    query.set(key, val);
    router.push(`${match.location.pathname}?${query.toString()}`);
  };

  requerySearchBox = filterValue => {
    this.setState({ filterValue });
    if (filterValue.length > 10) {
      this.changeQuery("filter", filterValue);
    }
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
              state={this.state.filterValue}
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
