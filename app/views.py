import json

from flask import render_template, request, jsonify
from app.basecoat import db_utils

from app import app, db, models



@app.route('/')
def index():
    formula_table = db_utils.get_table('Formula')
    formula_list = [formula for formula in formula_table]
    return render_template('index.html',
                           formula_list=formula_list)


@app.route('/formula/<int:formula_id>')
def get_formula(formula_id):
    formula = db_utils.get_object_from_table('Formula', 'id', formula_id)[0]
    colorant_list = db_utils.get_object_from_table('Colorant', 'formula_id', formula_id)
    base_list = db_utils.get_object_from_table('Base', 'formula_id', formula_id)
    return render_template('view_formula.html',
                           formula=formula,
                           colorant_list=colorant_list,
                           base_list=base_list)


@app.route('/formula/add', methods=['GET', 'POST'])
def add_formula():
    if request.method == 'POST':
        form_data = request.json
        colorant_data = form_data.pop('colorant_list', None)
        base_data = form_data.pop('base_list', None)

        form_data = {key: value.strip() for key, value in form_data.items()}

        if "formula_id" in form_data.keys():
            db_utils.update_db("Formula", "id", form_data['formula_id'], **form_data)
            for colorant in colorant_data:
                colorant_dict = {
                    'colorant_name': colorant,
                    'formula_id': form_data['formula_id'],
                    'amount': colorant_data[colorant]['colorant_amount'],
                }

                try:
                    colorant_dict['id'] = colorant_data[colorant]['colorant_id']
                except KeyError:
                    pass

                new_colorant = models.Colorant(**colorant_dict)

                try:
                    db.session.merge(new_colorant)
                    db.session.commit()
                except:
                    db.session.rollback()
                    raise

            for base in base_data:
                base_dict = {
                    'base_name': base,
                    'formula_id': form_data['formula_id'],
                    'product_name': base_data[base]['base_product_name'],
                }

                try:
                    base_dict['id'] = base_data[base]['base_id']
                except KeyError:
                    pass

                new_base = models.Base(**base_dict)

                try:
                    db.session.merge(new_base)
                    db.session.commit()
                except:
                    db.session.rollback()
                    raise

        else:
            new_formula = models.Formula(formula_name=form_data['formula_name'].title(),
                                         formula_number=form_data['formula_number'],
                                         customer_name=form_data['customer_name'].title(),
                                         summary=form_data['summary'],
                                         notes=form_data['notes'])
            db.session.add(new_formula)
            db.session.flush()

            for colorant in colorant_data:
                db_utils.insert_into_db('Colorant',
                               colorant_name=colorant,
                               formula_id=new_formula.id,
                               amount=colorant_data[colorant]['colorant_amount'])


            for base in base_data:
                db_utils.insert_into_db('Base',
                               base_name=base,
                               formula_id=new_formula.id,
                               product_name=base_data[base]['base_product_name'])

            try:
                db.session.commit()
            except:
                db.session.rollback()
                raise

        return jsonify({'success':True}), 200
    else:
        return render_template('add_formula.html')


@app.route('/formula/edit/<int:formula_id>')
def edit_formula(formula_id):
    formula = db_utils.get_object_from_table('Formula', 'id', formula_id)[0]
    colorant_list = db_utils.get_object_from_table('Colorant', 'formula_id', formula_id)
    base_list = db_utils.get_object_from_table('Base', 'formula_id', formula_id)
    return render_template('edit_formula.html',
                           formula=formula,
                           colorant_list=colorant_list,
                           base_list=base_list)
