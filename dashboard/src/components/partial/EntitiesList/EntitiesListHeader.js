import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Typography from "material-ui/Typography";
import Checkbox from "material-ui/Checkbox";
import { withStyles } from "material-ui/styles";

import { TableListHeader, TableListButton } from "/components/TableList";

import StatusMenu from "/components/partial/StatusMenu";

import EntitiesListSubscriptionsMenu from "./EntitiesListSubscriptionsMenu";

const styles = theme => ({
  headerButton: {
    marginLeft: theme.spacing.unit / 2,
    "&:first-child": {
      marginLeft: theme.spacing.unit,
    },
  },
  filterActions: {
    display: "none",
    [theme.breakpoints.up("sm")]: {
      display: "flex",
    },
  },
  // Remove padding from button container
  checkbox: {
    marginLeft: -11,
    color: theme.palette.primary.contrastText,
  },
  grow: {
    flex: "1 1 auto",
  },
});

class EntitiesListHeader extends React.PureComponent {
  static propTypes = {
    entities: PropTypes.object.isRequired,
    onChangeParams: PropTypes.func,
    onClickSelect: PropTypes.func,
    onClickSilence: PropTypes.func,
    selectedCount: PropTypes.number,
    classes: PropTypes.object.isRequired,
  };

  static defaultProps = {
    onChangeParams: () => {},
    onClickSelect: () => {},
    onClickSilence: () => {},
    selectedCount: 0,
  };

  static fragments = {
    entityConnection: gql`
      fragment EntitiesListHeader_entityConnection on EntityConnection {
        ...EntitiesListSubscriptionsMenu_entityConnection
      }

      ${EntitiesListSubscriptionsMenu.fragments.entityConnection}
    `,
  };

  render() {
    const {
      selectedCount,
      classes,
      onClickSelect,
      onClickSilence,
      onChangeParams,
      entities,
    } = this.props;

    return (
      <TableListHeader sticky active={selectedCount > 0}>
        <Checkbox
          component="button"
          className={classes.checkbox}
          onClick={onClickSelect}
          checked={false}
          indeterminate={selectedCount > 0}
        />
        {selectedCount > 0 && <div>{selectedCount} Selected</div>}
        <div className={classes.grow} />
        {selectedCount > 0 ? (
          <div>
            <TableListButton
              className={classes.headerButton}
              onClick={onClickSilence}
            >
              <Typography variant="button">Silence</Typography>
            </TableListButton>
          </div>
        ) : (
          <div className={classes.filterActions}>
            <EntitiesListSubscriptionsMenu
              entities={entities}
              value=""
              onChange={subscription => onChangeParams({ subscription })}
            />
            <StatusMenu onChange={status => onChangeParams({ status })} />
          </div>
        )}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
