import React from "react";
import PropTypes from "prop-types";
import Checkbox from "@material-ui/core/Checkbox";
import { withStyles } from "@material-ui/core/styles";

import Typography from "@material-ui/core/Typography";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemText from "@material-ui/core/ListItemText";

import {
  TableListHeader,
  TableListButton as Button,
  TableListSelect as Select,
} from "/components/TableList";
import ButtonSet from "/components/ButtonSet";

import ConfirmDelete from "/components/partials/ConfirmDelete";

const styles = theme => ({
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
    classes: PropTypes.object.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickDelete: PropTypes.func.isRequired,
    selectedCount: PropTypes.number.isRequired,
    onChangeQuery: PropTypes.func.isRequired,
  };

  _handleChangeSort = val => {
    let newVal = val;
    this.props.onChangeQuery(query => {
      // Toggle between ASC & DESC
      const curVal = query.get("order");
      if (curVal === "ID" && newVal === "ID") {
        newVal = "ID_DESC";
      }
      query.set("order", newVal);
    });
  };

  render() {
    const { classes, selectedCount, onClickSelect, onClickDelete } = this.props;

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
          <ButtonSet>
            <ConfirmDelete
              identifier={`${selectedCount} ${
                selectedCount === 1 ? "entity" : "entities"
              }`}
              onSubmit={onClickDelete}
            >
              {confirm => (
                <Button onClick={confirm.open}>
                  <Typography variant="button">Delete</Typography>
                </Button>
              )}
            </ConfirmDelete>
          </ButtonSet>
        ) : (
          <ButtonSet>
            <Select label="Sort" onChange={this._handleChangeSort}>
              <MenuItem key="ID" value="ID">
                <ListItemText>Name</ListItemText>
              </MenuItem>
              <MenuItem key="LASTSEEN" value="LASTSEEN">
                <ListItemText>Last Seen</ListItemText>
              </MenuItem>
            </Select>
          </ButtonSet>
        )}
      </TableListHeader>
    );
  }
}

export default withStyles(styles)(EntitiesListHeader);
