from flask_migrate import MigrateCommand
from faker import Faker
from app import db, models
import random
import json

fake = Faker()

@MigrateCommand.command
def populate_db(number_of_entries):
    "Populates database with fake data. Useful for developement"

    number_of_entries = int(number_of_entries)

    for entry in range(0, number_of_entries):
        dev_formula = models.Formula(formula_name=fake.color_name() + " " + fake.safe_color_name(),
                                     formula_number=fake.hex_color(),
                                     customer_name=fake.company(),
                                     colorants=json.dumps([[fake.color_name() + " " + fake.safe_color_name(), random.randint(1, 10)],
                                                           [fake.color_name() + " " + fake.safe_color_name(), random.randint(1, 10)]]),
                                     bases=json.dumps([[fake.color_name() + " " + fake.safe_color_name(), fake.company()]]),
                                     summary=fake.text(max_nb_chars=random.randint(50, 200)),
                                     notes=fake.paragraph(nb_sentences=3, variable_nb_sentences=True))

        try:
            db.session.add_all([dev_formula])
            db.session.commit()
        except:
            db.session.rollback()
            raise

    print('Populated formulas')


@MigrateCommand.command
def empty_db():
    """Clears all information from database"""
    meta = db.metadata

    for table in reversed(meta.sorted_tables):
        print('Cleared table: {}'.format(table))
        db.session.execute(table.delete())

    db.session.commit()
