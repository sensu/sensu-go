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

const defaultExpression = "HasCheck && IsIncident";

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
      $filter: String = "HasCheck && IsIncident"
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

    let filterValue = props.location.query.filter;
    if (filterValue === undefined) {
      filterValue = defaultExpression;
    }
    this.state = { filterValue };
  }

  changeQuery = (key, val) => {
    const { match, router } = this.props;
    const query = new URLSearchParams(match.location.query);

    if (key === "filter") {
      this.setState({ filterValue: val });
    }
    query.set(key, val);

    router.push(`${match.location.pathname}?${query.toString()}`);
  };

  requerySearchBox = filterValue => {
    if (filterValue.length >= 10) {
      this.changeQuery("filter", filterValue);
    }
    this.setState({ filterValue });
  };

  render() {
    const { classes, ...props } = this.props;
    return (
      <AppContent>
        <div>
          <div className={classes.headline}>
            <Typography className={classes.title} variant="headline">
              Recent Events
            </Typography>
            <SearchBox
              onUpdateInput={this.requerySearchBox}
              state={this.state.filterValue}
            />
          </div>
          <div className={classes.container}>
            <EventsContainer onQueryChange={this.changeQuery} {...props} />
          </div>
        </div>
      </AppContent>
    );
  }
}

export default withStyles(styles)(EventsPage);
