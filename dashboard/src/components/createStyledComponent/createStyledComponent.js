import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import {
  compose,
  setDisplayName,
  mapProps,
  defaultProps,
  componentFromProp,
} from "recompose";

import uniqueId from "/utils/uniqueId";

export const withStyle = styles => {
  const path = `with-style-${uniqueId()}`;
  return compose(
    withStyles(theme => ({ [path]: styles(theme) })),
    mapProps(
      ({ className, classes: { [path]: newClass, ...classes }, ...props }) => {
        const classesProps = {};
        if (classes.length > 0) {
          classesProps.classes = classes;
        }

        return {
          className: classnames(className, newClass),
          ...classesProps,
          ...props,
        };
      },
    ),
  );
};

const identity = cp => cp;
const componentFromComponentProp = componentFromProp("component");

// For ease of use, pre-built function for creating styled components.
export const createStyledComponent = ({ name, component = "div", styles }) => {
  const enhance = compose(
    defaultProps({ component }),
    withStyle(styles),
    name ? setDisplayName(name) : identity,
  );
  return enhance(componentFromComponentProp);
};

export default createStyledComponent;
